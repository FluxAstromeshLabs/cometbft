package hotstuff

import (
	"bytes"
	"context"
	"fmt"
	cs "github.com/cometbft/cometbft/consensus"
	hotstufftypes "github.com/cometbft/cometbft/consensus/hotstuff/types"
	"github.com/cometbft/cometbft/libs/fail"
	"github.com/cosmos/gogoproto/proto"
	"github.com/go-kit/kit/metrics/discard"
	"runtime/debug"
	"time"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	cmtevents "github.com/cometbft/cometbft/libs/events"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/types"

	privval "github.com/cometbft/cometbft/privval"
)

var msgQueueSize = 1000

// msgs from the reactor which may update the state
type msgInfo struct {
	Msg    Message `json:"msg"`
	PeerID p2p.ID  `json:"peer_key"`
}

// interface to the mempool
type txNotifier interface {
	TxsAvailable() <-chan struct{}
}

// interface to the evidence pool
type evidencePool interface {
	// reports conflicting votes to the evidence pool to be processed into evidence
	ReportConflictingVotes(voteA, voteB *types.Vote)
}

// State handles execution of the consensus algorithm.
// It processes votes and proposals, and upon reaching agreement,
// commits blocks to the chain and executes them against the application.
// The internal state machine receives input from peers, the internal validator, and from a timer.
type State struct {
	service.BaseService

	// config details
	config        *cfg.ConsensusConfig
	privValidator types.PrivValidator // for signing votes

	// store blocks and commits
	blockStore sm.BlockStore

	// create and execute blocks
	blockExec *sm.BlockExecutor

	// notify us if txs are available
	txNotifier txNotifier

	// add evidence to the pool
	// when it's detected
	evpool evidencePool

	// internal state
	mtx           cmtsync.RWMutex
	roundProgress *RoundProgress

	state sm.State // State until height-1.
	// privValidator pubkey, memoized for the duration of one block
	// to avoid extra requests to HSM
	privValidatorPubKey crypto.PubKey

	// state changes may be triggered by: msgs from peers,
	// msgs from ourself, or by timeouts
	peerMsgQueue     chan msgInfo
	internalMsgQueue chan msgInfo
	timeoutTicker    TimeoutTicker

	// information about about added votes and block parts are written on this channel
	// so statistics can be computed by reactor
	statsMsgQueue chan msgInfo

	// we use eventBus to trigger msg broadcasts in the reactor,
	// and to notify external subscribers, eg. through a websocket
	eventBus *types.EventBus

	// a Write-Ahead Log ensures we can recover from any kind of crash
	// and helps us avoid signing conflicting votes
	//wal          cs.WAL
	//replayMode   bool // so we don't log signing errors during replay
	//doWALCatchup bool // determines if we even try to do the catchup

	// for tests where we want to limit the number of transitions the state makes
	nSteps int

	// some functions can be overwritten for testing
	decideProposal func(height int64, round int32)
	doPrevote      func(height int64, round int32)
	setProposal    func(proposal *hotstufftypes.Proposal) error

	// closed when we finish shutting down
	done chan struct{}

	// synchronous pubsub between consensus state and reactor.
	// state only emits EventNewRoundStep and EventVote
	evsw cmtevents.EventSwitch

	// for reporting metrics
	metrics *cs.Metrics

	// offline state sync height indicating to which height the node synced offline
	offlineStateSyncHeight int64
}

// StateOption sets an optional parameter on the State.
type StateOption func(*State)

// NewState returns a new State.
func NewState(
	config *cfg.ConsensusConfig,
	state sm.State,
	blockExec *sm.BlockExecutor,
	blockStore sm.BlockStore,
	txNotifier txNotifier,
	evpool evidencePool,
	options ...StateOption,
) *State {
	s := &State{
		config:           config,
		blockExec:        blockExec,
		blockStore:       blockStore,
		txNotifier:       txNotifier,
		peerMsgQueue:     make(chan msgInfo, msgQueueSize),
		internalMsgQueue: make(chan msgInfo, msgQueueSize),
		timeoutTicker:    NewTimeoutTicker(),
		statsMsgQueue:    make(chan msgInfo, msgQueueSize),
		done:             make(chan struct{}),
		//doWALCatchup:     true,
		//wal:              cs.NilWAL{},
		evpool:  evpool,
		evsw:    cmtevents.NewEventSwitch(),
		metrics: NopMetrics(),
	}
	for _, option := range options {
		option(s)
	}
	// set function defaults (may be overwritten before calling Start)
	//s.decideProposal = s.defaultDecideProposal
	//s.doPrevote = s.defaultDoPrevote
	//s.setProposal = s.defaultSetProposal

	// We have no votes, so reconstruct LastCommit from SeenCommit.
	if state.LastBlockHeight > 0 {
		// In case of out of band performed statesync, the state store
		// will have a state but no extended commit (as no block has been downloaded).
		// If the height at which the vote extensions are enabled is lower
		// than the height at which we statesync, consensus will panic because
		// it will try to reconstruct the extended commit here.
		if s.offlineStateSyncHeight != 0 {
			s.reconstructSeenCommit(state)
		} else {
			s.reconstructLastCommit(state)
		}
	}

	// NOTE: we do not call scheduleRound0 yet, we do that upon Start()
	s.state = state

	s.BaseService = *service.NewBaseService(nil, "State", s)

	return s
}

// SetLogger implements Service.
func (s *State) SetLogger(l log.Logger) {
	s.BaseService.Logger = l
	s.timeoutTicker.SetLogger(l)
}

// SetEventBus sets event bus.
func (s *State) SetEventBus(b *types.EventBus) {
	s.eventBus = b
	s.blockExec.SetEventBus(b)
}

// StateMetrics sets the metrics.
func StateMetrics(metrics *cs.Metrics) StateOption {
	return func(s *State) { s.metrics = metrics }
}

// OfflineStateSyncHeight indicates the height at which the node
// statesync offline - before booting sets the metrics.
func OfflineStateSyncHeight(height int64) StateOption {
	return func(s *State) { s.offlineStateSyncHeight = height }
}

// String returns a string.
func (s *State) String() string {
	// better not to access shared variables
	return "ConsensusState"
}

// GetState returns a copy of the chain state.
func (s *State) GetState() sm.State {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.state.Copy()
}

// SetPrivValidator sets the private validator account for signing votes. It
// immediately requests pubkey and caches it.
func (s *State) SetPrivValidator(priv types.PrivValidator) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.privValidator = priv

	if err := s.updatePrivValidatorPubKey(); err != nil {
		s.Logger.Error("failed to get private validator pubkey", "err", err)
	}
}

// SetTimeoutTicker sets the local timer. It may be useful to overwrite for
// testing.
func (s *State) SetTimeoutTicker(timeoutTicker TimeoutTicker) {
	s.mtx.Lock()
	s.timeoutTicker = timeoutTicker
	s.mtx.Unlock()
}

// LoadCommit loads the commit for a given height.
func (s *State) LoadCommit(height int64) *types.Commit {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if height == s.blockStore.Height() {
		return s.blockStore.LoadSeenCommit(height)
	}

	return s.blockStore.LoadBlockCommit(height)
}

func (s *State) OnStart() error {
	fmt.Println("state OnStart")
	s.roundProgress = NewRoundProgress(s.state.Validators.Size(), 17, 0)

	// We may set the WAL in testing before calling Start, so only OpenWAL if its
	// still the nilWAL.
	//if _, ok := s.wal.(cs.NilWAL); ok {
	//	if err := s.loadWalFile(); err != nil {
	//		return err
	//	}
	//}

	// we need the timeoutRoutine for replay so
	// we don't block on the tick chan.
	// NOTE: we will get a build up of garbage go routines
	// firing on the tockChan until the receiveRoutine is started
	// to deal with them (by that point, at most one will be valid)
	if err := s.timeoutTicker.Start(); err != nil {
		return err
	}

	// We may have lost some votes if the process crashed reload from consensus
	// log to catchup.
	//if s.doWALCatchup {
	//	repairAttempted := false
	//
	//LOOP:
	//	for {
	//		err := s.catchupReplay(s.roundProgress.Height)
	//		switch {
	//		case err == nil:
	//			break LOOP
	//
	//		case !cs.IsDataCorruptionError(err):
	//			s.Logger.Error("error on catchup replay; proceeding to start state anyway", "err", err)
	//			break LOOP
	//
	//		case repairAttempted:
	//			return err
	//		}
	//
	//		s.Logger.Error("the WAL file is corrupted; attempting repair", "err", err)
	//
	//		// 1) prep work
	//		if err := s.wal.Stop(); err != nil {
	//			return err
	//		}
	//
	//		repairAttempted = true
	//
	//		// 2) backup original WAL file
	//		corruptedFile := fmt.Sprintf("%s.CORRUPTED", s.config.WalFile())
	//		if err := cmtos.CopyFile(s.config.WalFile(), corruptedFile); err != nil {
	//			return err
	//		}
	//
	//		s.Logger.Debug("backed up WAL file", "src", s.config.WalFile(), "dst", corruptedFile)
	//
	//		//// 3) try to repair (WAL file will be overwritten!)
	//		//if err := repairWalFile(corruptedFile, s.config.WalFile()); err != nil {
	//		//	s.Logger.Error("the WAL repair failed", "err", err)
	//		//	return err
	//		//}
	//
	//		s.Logger.Info("successful WAL repair")
	//
	//		// reload WAL file
	//		if err := s.loadWalFile(); err != nil {
	//			return err
	//		}
	//	}
	//}

	if err := s.evsw.Start(); err != nil {
		return err
	}

	// Double Signing Risk Reduction
	//if err := s.checkDoubleSigningRisk(s.Height); err != nil {
	//	return err
	//}

	// now start the receiveRoutine
	go s.receiveRoutine(0)

	// schedule the first round!
	// use GetRoundState so we don't race the receiveRoutine for access
	s.scheduleRound0()

	return nil
}

func (s *State) startRoutines(maxSteps int) {
	err := s.timeoutTicker.Start()
	if err != nil {
		s.Logger.Error("failed to start timeout ticker", "err", err)
		return
	}

	go s.receiveRoutine(maxSteps)
}

//func (s *State) loadWalFile() error {
//	wal, err := s.OpenWAL(s.config.WalFile())
//	if err != nil {
//		s.Logger.Error("failed to load state WAL", "err", err)
//		return err
//	}
//
//	s.wal = wal
//	return nil
//}

// OnStop implements service.Service.
func (s *State) OnStop() {
	if err := s.evsw.Stop(); err != nil {
		s.Logger.Error("failed trying to stop eventSwitch", "error", err)
	}

	if err := s.timeoutTicker.Stop(); err != nil {
		s.Logger.Error("failed trying to stop timeoutTicket", "error", err)
	}
	// WAL is stopped in receiveRoutine.
}

func (s *State) Wait() {
	<-s.done
}

func (s *State) OpenWAL(walFile string) (cs.WAL, error) {
	wal, err := cs.NewWAL(walFile)
	if err != nil {
		s.Logger.Error("failed to open WAL", "file", walFile, "err", err)
		return nil, err
	}

	wal.SetLogger(s.Logger.With("wal", walFile))

	if err := wal.Start(); err != nil {
		s.Logger.Error("failed to start WAL", "err", err)
		return nil, err
	}

	return wal, nil
}

func (s *State) AddVote(vote *types.Vote, peerID p2p.ID) (added bool, err error) {
	// TODO
	return false, nil
}

func (s *State) SetProposal(proposal *types.Proposal, peerID p2p.ID) error {
	// TODO
	return nil
}

func (s *State) AddProposalBlockPart(height int64, round int32, part *types.Part, peerID p2p.ID) error {
	// TODO
	return nil
}

func (s *State) SetProposalAndBlock(
	proposal *types.Proposal,
	block *types.Block, //nolint:revive
	parts *types.PartSet,
	peerID p2p.ID,
) error {
	// TODO
	return nil
}

func (s *State) updateHeight(height int64) {
	s.metrics.Height.Set(float64(height))
	s.roundProgress.Height = height
}

func (s *State) scheduleRound0() {
	fmt.Println("round 0 scheduled")

	addr := s.privValidatorPubKey.Address()
	if s.isProposer(addr) {
		s.scheduleTimeout(
			time.Millisecond*300,
			s.roundProgress.Height,
			s.roundProgress.Round,
			RoundStepLeaderPropose,
		)
	} else {
		s.scheduleTimeout(
			time.Millisecond*300,
			s.roundProgress.Height,
			s.roundProgress.Round,
			RoundStepValidatorPropose,
		)
	}
}

func (s *State) scheduleTimeout(duration time.Duration, height int64, round int32, step RoundStepType) {
	s.timeoutTicker.ScheduleTimeout(TimeoutInfo{duration, height, round, step})
}

func (s *State) sendInternalMessage(mi msgInfo) {
	select {
	case s.internalMsgQueue <- mi:
	default:
		// NOTE: using the go-routine means our votes can
		// be processed out of order.
		// TODO: use CList here for strict determinism and
		// attempt push to internalMsgQueue in receiveRoutine
		s.Logger.Debug("internal msg queue is full; using a go-routine")
		go func() { s.internalMsgQueue <- mi }()
	}
}

func (s *State) reconstructSeenCommit(state sm.State) {
	// TODO
}

func (s *State) reconstructLastCommit(state sm.State) {
	// TODO
}

func (s *State) votesFromExtendedCommit(state sm.State) (*types.VoteSet, error) {
	// TODO
	return nil, nil
}

func (s *State) votesFromSeenCommit(state sm.State) (*types.VoteSet, error) {
	// TODO
	return nil, nil
}

func (s *State) receiveRoutine(maxSteps int) {
	onExit := func(s *State) {
		// NOTE: the internalMsgQueue may have signed messages from our
		// priv_val that haven't hit the WAL, but its ok because
		// priv_val tracks LastSig

		// close wal now that we're done writing to it
		//if err := s.wal.Stop(); err != nil {
		//	s.Logger.Error("failed trying to stop WAL", "error", err)
		//}

		//s.wal.Wait()
		close(s.done)
	}

	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error("CONSENSUS FAILURE!!!", "err", r, "stack", string(debug.Stack()))
			// stop gracefully
			//
			// NOTE: We most probably shouldn't be running any further when there is
			// some unexpected panic. Some unknown error happened, and so we don't
			// know if that will result in the validator signing an invalid thing. It
			// might be worthwhile to explore a mechanism for manual resuming via
			// some console or secure RPC system, but for now, halting the chain upon
			// unexpected consensus bugs sounds like the better option.
			onExit(s)
		}
	}()

	for {
		if maxSteps > 0 {
			if s.nSteps >= maxSteps {
				s.Logger.Debug("reached max steps; exiting receive routine")
				s.nSteps = 0
				return
			}
		}

		var mi msgInfo

		select {
		case <-s.txNotifier.TxsAvailable():
			s.handleTxsAvailable()

		case mi = <-s.peerMsgQueue:
			//if err := s.wal.Write(mi); err != nil {
			//	s.Logger.Error("failed writing to WAL", "err", err)
			//}
			// handles proposals, block parts, votes
			// may generate internal events (votes, complete proposals, 2/3 majorities)
			s.handleMsg(mi)

		case mi = <-s.internalMsgQueue:
			//err := s.wal.WriteSync(mi) // NOTE: fsync
			//if err != nil {
			//	panic(fmt.Sprintf(
			//		"failed to write %v msg to consensus WAL due to %v; check your file system and restart the node",
			//		mi, err,
			//	))
			//}

			if _, ok := mi.Msg.(*VoteMessage); ok {
				// we actually want to simulate failing during
				// the previous WriteSync, but this isn't easy to do.
				// Equivalent would be to fail here and manually remove
				// some bytes from the end of the wal.
				fail.Fail() // XXX
			}

			// handles proposals, block parts, votes
			s.handleMsg(mi)

		case ti := <-s.timeoutTicker.Chan(): // tockChan:
			//if err := s.wal.Write(ti); err != nil {
			//	s.Logger.Error("failed writing to WAL", "err", err)
			//}

			// if the timeout is relevant to the rs
			// go to the next step
			s.handleTimeout(ti)

		case <-s.Quit():
			onExit(s)
			return
		}
	}
}

func (s *State) handleMsg(mi msgInfo) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	var err error

	msg, peerID := mi.Msg, mi.PeerID
	addr := s.privValidatorPubKey.Address()
	isLeader := s.isProposer(addr)

	// leader msg handlers
	if isLeader {
		switch msg := msg.(type) {
		case *hotstufftypes.Vote:
			fmt.Println("receive msg vote", msg, peerID)
			switch msg.Type {
			case hotstufftypes.PrepareVote:
				addr := s.privValidatorPubKey.Address()
				if s.isProposer(addr) {
					valIdx, _ := s.state.Validators.GetByAddress(msg.Address)
					s.roundProgress.PrepareQC.SetVote(msg.Signature, valIdx)
				}
				fmt.Println("prepare QC", s.roundProgress.PrepareQC)
			}
		}
	}

	// validators msg handlers
	if !isLeader {
		switch msg := msg.(type) {
		case *hotstufftypes.Proposal:
			// sign and vote
			fmt.Println("receive msg proposal", msg, peerID)
			vote, e := s.signVoteProposal(msg)
			err = e
			fmt.Println(fmt.Sprintf("validator broadcasting %s", VoteEvent))
			s.evsw.FireEvent(VoteEvent, vote)

			// set proposal
			s.roundProgress.Proposal = *msg

		case *hotstufftypes.QuorumCert:
			fmt.Println("received quorum cert", msg)

		default:
			s.Logger.Error("unknown msg type", "type", fmt.Sprintf("%T", msg))
			return
		}
	}

	if err != nil {
		s.Logger.Error(
			"failed to process message",
			"height", s.roundProgress.Height,
			"round", s.roundProgress.Round,
			"peer", peerID,
			"msg_type", fmt.Sprintf("%T", msg),
			"err", err,
		)
	}
}

func (s *State) handleTimeout(ti TimeoutInfo) {
	s.Logger.Debug("hotstuff timeout", "timeout", ti.Duration, "height", ti.Height, "round", ti.Round, "step", ti.Step)

	s.mtx.Lock()
	defer s.mtx.Unlock()

	addr := s.privValidatorPubKey.Address()
	isLeader := s.isProposer(addr)

	// leader timeout handlers
	if isLeader {
		switch ti.Step {
		case RoundStepLeaderPropose:
			if s.privValidatorPubKey == nil {
				return
			}
			if !s.state.Validators.HasAddress(addr) {
				return
			}
			fmt.Println(fmt.Sprintf("proposer %s broadcasting %s", addr.String(), ProposalEvent))
			proposal := &hotstufftypes.Proposal{Height: ti.Height, Round: ti.Round}
			s.evsw.FireEvent(ProposalEvent, proposal)

			// proposer votes
			vote, err := s.signVoteProposal(proposal)
			if err != nil {
				s.Logger.Error("failed to sign vote", "err", err)
			}
			proposerIdx, _ := s.state.Validators.GetByAddress(addr)
			s.roundProgress.PrepareQC.SetVote(vote.Signature, proposerIdx)

			// start timeout for next step
			s.scheduleTimeout(
				time.Millisecond*300,
				s.roundProgress.Height,
				s.roundProgress.Round,
				RoundStepLeaderPrepare,
			)

		case RoundStepLeaderPrepare:
			if s.privValidatorPubKey == nil {
				return
			}
			if !s.state.Validators.HasAddress(addr) {
				return
			}

			// check if leader collects enough votes
			validCert := s.roundProgress.PrepareQC.HasQuorum()
			if validCert {
				// broadcast prepare QC to validators
				fmt.Println("prepare QC is valid, broadcasting to validators")
				s.evsw.FireEvent(QCEvent, &s.roundProgress.PrepareQC)
			} else {
				// act as validator and cast gossip for view-change QC
				fmt.Println("prepare QC didn't have quorum, starting gossip for view-change")
			}
		}
	}

	// validators timeout handlers
	if !isLeader {
		switch ti.Step {
		case RoundStepValidatorPropose:
			// TODO: if proposal not received, gossip for view-change QC

		case RoundStepValidatorPrepare:
			// TODO: if prepare QC not received, gossip for view-change QC
		}
	}

}

func (s *State) signVoteProposal(msg *hotstufftypes.Proposal) (*hotstufftypes.Vote, error) {
	pv, ok := s.privValidator.(*privval.FilePV)
	if !ok {
		panic("cannot cast priv validator key to concrete type")
	}

	addr := s.privValidatorPubKey.Address()
	vote := &hotstufftypes.Vote{
		Type:      hotstufftypes.PrepareVote,
		Address:   addr,
		Signature: nil,
	}

	signBytes, err := proto.Marshal(vote)
	if err != nil {
		return nil, err
	}

	sig, err := pv.Key.PrivKey.Sign(signBytes)
	if err != nil {
		return nil, err
	}

	vote.Signature = sig

	return vote, nil
}

// ---------------------------------------------------------------

func (s *State) handleTxsAvailable() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
}

func (s *State) isProposer(address []byte) bool {
	return bytes.Equal(s.state.Validators.GetProposer().Address, address)
}

func (s *State) createProposalBlock(ctx context.Context) (*types.Block, error) {
	return &types.Block{}, nil
}

func (s *State) signVote(
	msgType cmtproto.SignedMsgType,
	hash []byte,
	header types.PartSetHeader,
	block *types.Block,
) (*types.Vote, error) {
	return &types.Vote{}, nil
}

func (s *State) updatePrivValidatorPubKey() error {
	if s.privValidator == nil {
		return nil
	}

	pubKey, err := s.privValidator.GetPubKey()
	if err != nil {
		return err
	}
	s.privValidatorPubKey = pubKey
	return nil
}

func NopMetrics() *cs.Metrics {
	return &cs.Metrics{
		Height:                    discard.NewGauge(),
		ValidatorLastSignedHeight: discard.NewGauge(),
		Rounds:                    discard.NewGauge(),
		RoundDurationSeconds:      discard.NewHistogram(),
		Validators:                discard.NewGauge(),
		ValidatorsPower:           discard.NewGauge(),
		ValidatorPower:            discard.NewGauge(),
		ValidatorMissedBlocks:     discard.NewGauge(),
		MissingValidators:         discard.NewGauge(),
		MissingValidatorsPower:    discard.NewGauge(),
		ByzantineValidators:       discard.NewGauge(),
		ByzantineValidatorsPower:  discard.NewGauge(),
		BlockIntervalSeconds:      discard.NewHistogram(),
		NumTxs:                    discard.NewGauge(),
		BlockSizeBytes:            discard.NewGauge(),
		//ChainSizeBytes:            discard.NewCounter(),
		TotalTxs:                  discard.NewGauge(),
		CommittedHeight:           discard.NewGauge(),
		BlockParts:                discard.NewCounter(),
		DuplicateBlockPart:        discard.NewCounter(),
		DuplicateVote:             discard.NewCounter(),
		StepDurationSeconds:       discard.NewHistogram(),
		BlockGossipPartsReceived:  discard.NewCounter(),
		QuorumPrevoteDelay:        discard.NewGauge(),
		FullPrevoteDelay:          discard.NewGauge(),
		VoteExtensionReceiveCount: discard.NewCounter(),
		ProposalReceiveCount:      discard.NewCounter(),
		ProposalCreateCount:       discard.NewCounter(),
		RoundVotingPowerPercent:   discard.NewGauge(),
		LateVotes:                 discard.NewCounter(),
	}
}

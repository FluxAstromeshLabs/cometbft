package hotstuff

import (
	"bytes"
	"context"
	"fmt"
	cs "github.com/cometbft/cometbft/consensus"
	hotstufftypes "github.com/cometbft/cometbft/consensus/hotstuff/types"
	"github.com/cometbft/cometbft/libs/fail"
	cmtmath "github.com/cometbft/cometbft/libs/math"
	cmtos "github.com/cometbft/cometbft/libs/os"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/go-kit/kit/metrics/discard"
	"os"
	"runtime/debug"
	"sort"
	"time"

	cfg "github.com/cometbft/cometbft/config"
	cstypes "github.com/cometbft/cometbft/consensus/types"
	"github.com/cometbft/cometbft/crypto"
	cmtevents "github.com/cometbft/cometbft/libs/events"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/types"
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
	mtx cmtsync.RWMutex
	cstypes.RoundState
	state sm.State // State until height-1.
	// privValidator pubkey, memoized for the duration of one block
	// to avoid extra requests to HSM
	privValidatorPubKey crypto.PubKey

	// state changes may be triggered by: msgs from peers,
	// msgs from ourself, or by timeouts
	peerMsgQueue     chan msgInfo
	internalMsgQueue chan msgInfo
	timeoutTicker    cs.TimeoutTicker

	// information about about added votes and block parts are written on this channel
	// so statistics can be computed by reactor
	statsMsgQueue chan msgInfo

	// we use eventBus to trigger msg broadcasts in the reactor,
	// and to notify external subscribers, eg. through a websocket
	eventBus *types.EventBus

	// a Write-Ahead Log ensures we can recover from any kind of crash
	// and helps us avoid signing conflicting votes
	wal          cs.WAL
	replayMode   bool // so we don't log signing errors during replay
	doWALCatchup bool // determines if we even try to do the catchup

	// for tests where we want to limit the number of transitions the state makes
	nSteps int

	// some functions can be overwritten for testing
	decideProposal func(height int64, round int32)
	doPrevote      func(height int64, round int32)
	setProposal    func(proposal *types.Proposal) error

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
		timeoutTicker:    cs.NewTimeoutTicker(),
		statsMsgQueue:    make(chan msgInfo, msgQueueSize),
		done:             make(chan struct{}),
		doWALCatchup:     true,
		wal:              cs.NilWAL{},
		evpool:           evpool,
		evsw:             cmtevents.NewEventSwitch(),
		metrics:          NopMetrics(),
	}
	for _, option := range options {
		option(s)
	}
	// set function defaults (may be overwritten before calling Start)
	s.decideProposal = s.defaultDecideProposal
	s.doPrevote = s.defaultDoPrevote
	s.setProposal = s.defaultSetProposal

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

	s.updateToState(state)

	// NOTE: we do not call scheduleRound0 yet, we do that upon Start()

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

// GetLastHeight returns the last height committed.
// If there were no blocks, returns 0.
func (s *State) GetLastHeight() int64 {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.RoundState.Height - 1
}

// GetRoundState returns a shallow copy of the internal consensus state.
func (s *State) GetRoundState() *cstypes.RoundState {
	s.mtx.RLock()
	rs := s.RoundState // copy
	s.mtx.RUnlock()
	return &rs
}

// GetRoundStateJSON returns a json of RoundState.
func (s *State) GetRoundStateJSON() ([]byte, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return cmtjson.Marshal(s.RoundState)
}

// GetRoundStateSimpleJSON returns a json of RoundStateSimple
func (s *State) GetRoundStateSimpleJSON() ([]byte, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return cmtjson.Marshal(s.RoundState.RoundStateSimple())
}

// GetValidators returns a copy of the current validators.
func (s *State) GetValidators() (int64, []*types.Validator) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.state.LastBlockHeight, s.state.Validators.Copy().Validators
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
func (s *State) SetTimeoutTicker(timeoutTicker cs.TimeoutTicker) {
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

	// We may set the WAL in testing before calling Start, so only OpenWAL if its
	// still the nilWAL.
	if _, ok := s.wal.(cs.NilWAL); ok {
		if err := s.loadWalFile(); err != nil {
			return err
		}
	}

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
	if s.doWALCatchup {
		repairAttempted := false

	LOOP:
		for {
			err := s.catchupReplay(s.Height)
			switch {
			case err == nil:
				break LOOP

			case !cs.IsDataCorruptionError(err):
				s.Logger.Error("error on catchup replay; proceeding to start state anyway", "err", err)
				break LOOP

			case repairAttempted:
				return err
			}

			s.Logger.Error("the WAL file is corrupted; attempting repair", "err", err)

			// 1) prep work
			if err := s.wal.Stop(); err != nil {
				return err
			}

			repairAttempted = true

			// 2) backup original WAL file
			corruptedFile := fmt.Sprintf("%s.CORRUPTED", s.config.WalFile())
			if err := cmtos.CopyFile(s.config.WalFile(), corruptedFile); err != nil {
				return err
			}

			s.Logger.Debug("backed up WAL file", "src", s.config.WalFile(), "dst", corruptedFile)

			// 3) try to repair (WAL file will be overwritten!)
			if err := repairWalFile(corruptedFile, s.config.WalFile()); err != nil {
				s.Logger.Error("the WAL repair failed", "err", err)
				return err
			}

			s.Logger.Info("successful WAL repair")

			// reload WAL file
			if err := s.loadWalFile(); err != nil {
				return err
			}
		}
	}

	if err := s.evsw.Start(); err != nil {
		return err
	}

	// Double Signing Risk Reduction
	if err := s.checkDoubleSigningRisk(s.Height); err != nil {
		return err
	}

	// now start the receiveRoutine
	go s.receiveRoutine(0)

	// schedule the first round!
	// use GetRoundState so we don't race the receiveRoutine for access
	s.scheduleRound0(s.GetRoundState())

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

func (s *State) loadWalFile() error {
	wal, err := s.OpenWAL(s.config.WalFile())
	if err != nil {
		s.Logger.Error("failed to load state WAL", "err", err)
		return err
	}

	s.wal = wal
	return nil
}

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
	s.Height = height
}

func (s *State) updateRoundStep(round int32, step cstypes.RoundStepType) {
	if !s.replayMode {
		if round != s.Round || round == 0 && step == cstypes.RoundStepNewRound {
			s.metrics.MarkRound(s.Round, s.StartTime)
		}
		if s.Step != step {
			s.metrics.MarkStep(s.Step)
		}
	}
	s.Round = round
	s.Step = step
}

func (s *State) scheduleRound0(rs *cstypes.RoundState) {
	fmt.Println("round 0 scheduled")
	sleepDuration := time.Millisecond * 300
	s.scheduleTimeout(sleepDuration, 17, 0, cstypes.RoundStepNewHeight)
}

func (s *State) scheduleTimeout(duration time.Duration, height int64, round int32, step cstypes.RoundStepType) {
	s.timeoutTicker.ScheduleTimeout(cs.TimeoutInfo{duration, height, round, step})
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

func (s *State) updateToState(state sm.State) {
	validators := state.Validators
	s.Validators = validators
	s.state = state

}

func (s *State) newStep() {
	rs := s.RoundStateEvent()
	if err := s.wal.Write(rs); err != nil {
		s.Logger.Error("failed writing to WAL", "err", err)
	}

	s.nSteps++

	// newStep is called by updateToState in NewState before the eventBus is set!
	if s.eventBus != nil {
		if err := s.eventBus.PublishEventNewRoundStep(rs); err != nil {
			s.Logger.Error("failed publishing new round step", "err", err)
		}

		s.evsw.FireEvent(types.EventNewRoundStep, &s.RoundState)
	}
}

func (s *State) receiveRoutine(maxSteps int) {
	onExit := func(s *State) {
		// NOTE: the internalMsgQueue may have signed messages from our
		// priv_val that haven't hit the WAL, but its ok because
		// priv_val tracks LastSig

		// close wal now that we're done writing to it
		if err := s.wal.Stop(); err != nil {
			s.Logger.Error("failed trying to stop WAL", "error", err)
		}

		s.wal.Wait()
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

		rs := s.RoundState
		var mi msgInfo

		select {
		case <-s.txNotifier.TxsAvailable():
			s.handleTxsAvailable()

		case mi = <-s.peerMsgQueue:
			if err := s.wal.Write(mi); err != nil {
				s.Logger.Error("failed writing to WAL", "err", err)
			}
			// handles proposals, block parts, votes
			// may generate internal events (votes, complete proposals, 2/3 majorities)
			s.handleMsg(mi)

		case mi = <-s.internalMsgQueue:
			err := s.wal.WriteSync(mi) // NOTE: fsync
			if err != nil {
				panic(fmt.Sprintf(
					"failed to write %v msg to consensus WAL due to %v; check your file system and restart the node",
					mi, err,
				))
			}

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
			if err := s.wal.Write(ti); err != nil {
				s.Logger.Error("failed writing to WAL", "err", err)
			}

			// if the timeout is relevant to the rs
			// go to the next step
			s.handleTimeout(ti, rs)

		case <-s.Quit():
			onExit(s)
			return
		}
	}
}

func (s *State) handleMsg(mi msgInfo) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	var (
		added bool
		err   error
	)

	msg, peerID := mi.Msg, mi.PeerID

	switch msg := msg.(type) {
	case *ProposalMessage:
		// will not cause transition.
		// once proposal is set, we can receive block parts
		err = s.setProposal(msg.Proposal)

	case *BlockPartMessage:
		// if the proposal is complete, we'll enterPrevote or tryFinalizeCommit
		added, err = s.addProposalBlockPart(msg, peerID)

		// We unlock here to yield to any routines that need to read the the RoundState.
		// Previously, this code held the lock from the point at which the final block
		// part was received until the block executed against the application.
		// This prevented the reactor from being able to retrieve the most updated
		// version of the RoundState. The reactor needs the updated RoundState to
		// gossip the now completed block.
		//
		// This code can be further improved by either always operating on a copy
		// of RoundState and only locking when switching out State's copy of
		// RoundState with the updated copy or by emitting RoundState events in
		// more places for routines depending on it to listen for.
		s.mtx.Unlock()

		s.mtx.Lock()
		if added && s.ProposalBlockParts.IsComplete() {
			s.handleCompleteProposal(msg.Height)
		}
		if added {
			s.statsMsgQueue <- mi
		}

		if err != nil && msg.Round != s.Round {
			s.Logger.Debug(
				"received block part from wrong round",
				"height", s.Height,
				"cs_round", s.Round,
				"block_round", msg.Round,
			)
			err = nil
		}

	case *VoteMessage:
		// attempt to add the vote and dupeout the validator if its a duplicate signature
		// if the vote gives us a 2/3-any or 2/3-one, we transition
		added, err = s.tryAddVote(msg.Vote, peerID)
		if added {
			s.statsMsgQueue <- mi
		}

		// if err == ErrAddingVote {
		// TODO: punish peer
		// We probably don't want to stop the peer here. The vote does not
		// necessarily comes from a malicious peer but can be just broadcasted by
		// a typical peer.
		// https://github.com/tendermint/tendermint/issues/1281
		// }

		// NOTE: the vote is broadcast to peers by the reactor listening
		// for vote events

		// TODO: If rs.Height == vote.Height && rs.Round < vote.Round,
		// the peer is sending us CatchupCommit precommits.
		// We could make note of this and help filter in broadcastHasVoteMessage().

	default:
		s.Logger.Error("unknown msg type", "type", fmt.Sprintf("%T", msg))
		return
	}

	if err != nil {
		s.Logger.Error(
			"failed to process message",
			"height", s.Height,
			"round", s.Round,
			"peer", peerID,
			"msg_type", fmt.Sprintf("%T", msg),
			"err", err,
		)
	}
}

func (s *State) handleTimeout(ti cs.TimeoutInfo, rs cstypes.RoundState) {
	s.Logger.Debug("hotstuff timeout", "timeout", ti.Duration, "height", ti.Height, "round", ti.Round, "step", ti.Step)

	// the timeout will now cause a state transition
	s.mtx.Lock()
	defer s.mtx.Unlock()

	switch ti.Step {
	default:
		// only fire event if this is proposer
		if s.privValidatorPubKey == nil {
			return
		}
		address := s.privValidatorPubKey.Address()
		if !s.Validators.HasAddress(address) {
			return
		}
		if s.isProposer(address) {
			fmt.Println(fmt.Sprintf("proposer %s broadcasting %s", address.String(), ProposalEvent))
			s.evsw.FireEvent(ProposalEvent, &hotstufftypes.Proposal{Type: hotstufftypes.ProposalType, Height: ti.Height, Round: ti.Round})
		}

	}
}

func (s *State) handleTxsAvailable() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// We only need to do this for round 0.
	if s.Round != 0 {
		return
	}

	switch s.Step {
	case cstypes.RoundStepNewHeight: // timeoutCommit phase
		if s.needProofBlock(s.Height) {
			// enterPropose will be called by enterNewRound
			return
		}

		// +1ms to ensure RoundStepNewRound timeout always happens after RoundStepNewHeight
		timeoutCommit := s.StartTime.Sub(cmttime.Now()) + 1*time.Millisecond
		s.scheduleTimeout(timeoutCommit, s.Height, 0, cstypes.RoundStepNewRound)

	case cstypes.RoundStepNewRound: // after timeoutCommit
		s.enterPropose(s.Height, 0)
	}
}

func (s *State) enterNewRound(height int64, round int32) {
	// TODO: learn what happened here and port hotstuff
	fmt.Println("enterNewRound", height, round)

	logger := s.Logger.With("height", height, "round", round)

	if s.Height != height || round < s.Round || (s.Round == round && s.Step != cstypes.RoundStepNewHeight) {
		logger.Debug(
			"entering new round with invalid args",
			"current", log.NewLazySprintf("%v/%v/%v", s.Height, s.Round, s.Step),
		)
		return
	}

	if now := cmttime.Now(); s.StartTime.After(now) {
		logger.Debug("need to set a buffer and log message here for sanity", "start_time", s.StartTime, "now", now)
	}

	prevHeight, prevRound, prevStep := s.Height, s.Round, s.Step

	// increment validators if necessary
	validators := s.Validators
	if s.Round < round {
		validators = validators.Copy()
		validators.IncrementProposerPriority(cmtmath.SafeSubInt32(round, s.Round))
	}

	// Setup new round
	// we don't fire newStep for this step,
	// but we fire an event, so update the round step first
	s.updateRoundStep(round, cstypes.RoundStepNewRound)
	s.Validators = validators
	// If round == 0, we've already reset these upon new height, and meanwhile
	// we might have received a proposal for round 0.
	propAddress := validators.GetProposer().PubKey.Address()
	if round != 0 {
		logger.Info("resetting proposal info", "proposer", propAddress)
		s.Proposal = nil
		s.ProposalBlock = nil
		s.ProposalBlockParts = nil
	}

	logger.Debug("entering new round",
		"previous", log.NewLazySprintf("%v/%v/%v", prevHeight, prevRound, prevStep),
		"proposer", propAddress,
	)

	s.Votes.SetRound(cmtmath.SafeAddInt32(round, 1)) // also track next round (round+1) to allow round-skipping
	s.TriggeredTimeoutPrecommit = false

	if err := s.eventBus.PublishEventNewRound(s.NewRoundEvent()); err != nil {
		s.Logger.Error("failed publishing new round", "err", err)
	}

	// Wait for txs to be available in the mempool
	// before we enterPropose in round 0. If the last block changed the app hash,
	// we may need an empty "proof" block, and enterPropose immediately.
	waitForTxs := s.config.WaitForTxs() && round == 0 && !s.needProofBlock(height)
	if waitForTxs {
		if s.config.CreateEmptyBlocksInterval > 0 {
			s.scheduleTimeout(s.config.CreateEmptyBlocksInterval, height, round,
				cstypes.RoundStepNewRound)
		}
	} else {
		s.enterPropose(height, round)
	}
}

func (s *State) needProofBlock(height int64) bool {
	return false
}

func (s *State) enterPropose(height int64, round int32) {
}

func (s *State) isProposer(address []byte) bool {
	return bytes.Equal(s.Validators.GetProposer().Address, address)
}

func (s *State) defaultDecideProposal(height int64, round int32) {
}

func (s *State) isProposalComplete() bool {
	if s.Proposal == nil || s.ProposalBlock == nil {
		return false
	}
	// we have the proposal. if there's a POLRound,
	// make sure we have the prevotes from it too
	if s.Proposal.POLRound < 0 {
		return true
	}
	// if this is false the proposer is lying or we haven't received the POL yet
	return s.Votes.Prevotes(s.Proposal.POLRound).HasTwoThirdsMajority()
}

func (s *State) createProposalBlock(ctx context.Context) (*types.Block, error) {
	return &types.Block{}, nil
}

func (s *State) enterPrevote(height int64, round int32) {
}

func (s *State) defaultDoPrevote(height int64, round int32) {
}

func (s *State) enterPrevoteWait(height int64, round int32) {
}

func (s *State) enterPrecommit(height int64, round int32) {
}

func (s *State) enterPrecommitWait(height int64, round int32) {
}

func (s *State) enterCommit(height int64, commitRound int32) {
}

func (s *State) tryFinalizeCommit(height int64) {
}

func (s *State) finalizeCommit(height int64) {
}

func (s *State) recordMetrics(height int64, block *types.Block) {
}

func (s *State) defaultSetProposal(proposal *types.Proposal) error {
	return nil
}

func (s *State) addProposalBlockPart(msg *BlockPartMessage, peerID p2p.ID) (added bool, err error) {
	return true, nil
}

func (s *State) handleCompleteProposal(blockHeight int64) {
}

func (s *State) tryAddVote(vote *types.Vote, peerID p2p.ID) (bool, error) {
	return true, nil
}

func (s *State) addVote(vote *types.Vote, peerID p2p.ID) (added bool, err error) {
	return true, nil
}

func (s *State) signVote(
	msgType cmtproto.SignedMsgType,
	hash []byte,
	header types.PartSetHeader,
	block *types.Block,
) (*types.Vote, error) {
	return &types.Vote{}, nil
}

func (s *State) voteTime() time.Time {
	return time.Time{}
}

func (s *State) signAddVote(
	msgType cmtproto.SignedMsgType,
	hash []byte,
	header types.PartSetHeader,
	block *types.Block,
) {
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

func (s *State) checkDoubleSigningRisk(height int64) error {
	return nil
}

func (s *State) calculatePrevoteMessageDelayMetrics() {
	if s.Proposal == nil {
		return
	}

	ps := s.Votes.Prevotes(s.Round)
	pl := ps.List()

	sort.Slice(pl, func(i, j int) bool {
		return pl[i].Timestamp.Before(pl[j].Timestamp)
	})

	var votingPowerSeen int64
	for _, v := range pl {
		_, val := s.Validators.GetByAddress(v.ValidatorAddress)
		votingPowerSeen += val.VotingPower
		if votingPowerSeen >= s.Validators.TotalVotingPower()*2/3+1 {
			s.metrics.QuorumPrevoteDelay.With("proposer_address", s.Validators.GetProposer().Address.String()).Set(v.Timestamp.Sub(s.Proposal.Timestamp).Seconds())
			break
		}
	}
	if ps.HasAll() {
		s.metrics.FullPrevoteDelay.With("proposer_address", s.Validators.GetProposer().Address.String()).Set(pl[len(pl)-1].Timestamp.Sub(s.Proposal.Timestamp).Seconds())
	}
}

//---------------------------------------------------------

func CompareHRS(h1 int64, r1 int32, s1 cstypes.RoundStepType, h2 int64, r2 int32, s2 cstypes.RoundStepType) int {
	if h1 < h2 {
		return -1
	} else if h1 > h2 {
		return 1
	}
	if r1 < r2 {
		return -1
	} else if r1 > r2 {
		return 1
	}
	if s1 < s2 {
		return -1
	} else if s1 > s2 {
		return 1
	}
	return 0
}

// repairWalFile decodes messages from src (until the decoder errors) and
// writes them to dst.
func repairWalFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	var (
		dec = cs.NewWALDecoder(in)
		enc = cs.NewWALEncoder(out)
	)

	// best-case repair (until first error is encountered)
	for {
		msg, err := dec.Decode()
		if err != nil {
			break
		}

		err = enc.Encode(msg)
		if err != nil {
			return fmt.Errorf("failed to encode msg: %w", err)
		}
	}

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

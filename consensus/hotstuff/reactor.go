package hotstuff

import (
	"errors"
	"fmt"
	cs "github.com/cometbft/cometbft/consensus"
	hotstufftypes "github.com/cometbft/cometbft/consensus/hotstuff/types"
	cstypes "github.com/cometbft/cometbft/consensus/types"
	"github.com/cometbft/cometbft/libs/bits"
	cmtevents "github.com/cometbft/cometbft/libs/events"
	"github.com/cometbft/cometbft/libs/log"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	cmtcons "github.com/cometbft/cometbft/proto/tendermint/consensus"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/types"
	"sync"
)

type Message interface {
	ValidateBasic() error
}

var (
	_ p2p.Reactor = &HotstuffReactor{}
)

type HotstuffReactor struct {
	p2p.BaseReactor // BaseService + p2p.Switch

	conS *State

	broadcastFunc func(e p2p.Envelope) chan bool

	mtx      cmtsync.RWMutex
	waitSync bool
	eventBus *types.EventBus
	rs       *cstypes.RoundState

	Metrics *cs.Metrics
}

func NewReactor(consensusState *State) *HotstuffReactor {
	conR := &HotstuffReactor{
		conS:     consensusState,
		waitSync: false,
		rs:       consensusState.GetRoundState(),
		Metrics:  cs.NopMetrics(),
	}
	conR.BaseReactor = *p2p.NewBaseReactor("Consensus", conR)

	conR.broadcastFunc = func(e p2p.Envelope) chan bool {
		conR.Switch.Logger.Debug("Hotstuff Broadcast", "channel", e.ChannelID)

		peers := conR.Switch.Peers().List()
		var wg sync.WaitGroup
		wg.Add(len(peers))
		successChan := make(chan bool, len(peers))

		for _, peer := range peers {
			go func(p p2p.Peer) {
				defer wg.Done()
				success := p.Send(e)
				successChan <- success
			}(peer)
		}

		go func() {
			wg.Wait()
			close(successChan)
		}()

		return successChan
	}

	//for _, option := range options {
	//	option(conR)
	//}

	return conR
}

func (conR *HotstuffReactor) GetReactors() map[string]p2p.Reactor {
	return map[string]p2p.Reactor{
		"HOTSTUFF_TEST": conR,
	}
}

func (conR *HotstuffReactor) OnStart() error {
	conR.Logger.Info("Reactor ", "waitSync", conR.WaitSync())

	// start routine that computes peer statistics for evaluating peer quality
	go conR.peerStatsRoutine()

	conR.subscribeToBroadcastEvents()

	// get state and set to conR.rs every 100 micro second, inefficient
	go conR.updateRoundStateRoutine()

	if !conR.WaitSync() {
		err := conR.conS.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (conR *HotstuffReactor) OnStop() {
	conR.unsubscribeFromBroadcastEvents()
}

func (conR *HotstuffReactor) SwitchToConsensus(state sm.State, skipWAL bool) {
}

func (conR *HotstuffReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  ProposalChannel,
			Priority:            6,
			SendQueueCapacity:   100,
			RecvMessageCapacity: MaxMsgSize,
			MessageType:         &hotstufftypes.Proposal{},
		},
		{
			ID:                  BarChannel,
			Priority:            10,
			SendQueueCapacity:   100,
			RecvBufferCapacity:  50 * 4096,
			RecvMessageCapacity: MaxMsgSize,
			MessageType:         &cmtcons.Message{},
		},
		{
			ID:                  PingChannel,
			Priority:            7,
			SendQueueCapacity:   100,
			RecvBufferCapacity:  100 * 100,
			RecvMessageCapacity: MaxMsgSize,
			MessageType:         &cmtcons.Message{},
		},
		{
			ID:                  PongChannel,
			Priority:            1,
			SendQueueCapacity:   2,
			RecvBufferCapacity:  1024,
			RecvMessageCapacity: MaxMsgSize,
			MessageType:         &cmtcons.Message{},
		},
	}
}

func (conR *HotstuffReactor) InitPeer(peer p2p.Peer) p2p.Peer {
	peerState := cs.NewPeerState(peer).SetLogger(conR.Logger)
	peer.Set(types.PeerStateKey, peerState)
	return peer
}

func (conR *HotstuffReactor) AddPeer(peer p2p.Peer) {
}

func (conR *HotstuffReactor) RemovePeer(p2p.Peer, interface{}) {
}

func (conR *HotstuffReactor) Receive(e p2p.Envelope) {
	ps, ok := e.Src.Get(types.PeerStateKey).(*cs.PeerState)
	if !ok {
		panic(fmt.Sprintf("Peer %v has no state", e.Src))
	}

	switch e.ChannelID {
	case ProposalChannel:
		switch msg := e.Message.(type) {
		case *hotstufftypes.Proposal:
			fmt.Println("Receive", ps.String(), msg)
		}
	}
}

func (conR *HotstuffReactor) SetEventBus(b *types.EventBus) {
	conR.eventBus = b
	conR.conS.SetEventBus(b)
}

func (conR *HotstuffReactor) WaitSync() bool {
	return false
}

func (conR *HotstuffReactor) subscribeToBroadcastEvents() {
	const subscriber = "hotstuff-reactor"
	if err := conR.conS.evsw.AddListenerForEvent(subscriber, ProposalEvent,
		func(data cmtevents.EventData) {
			conR.broadcastProposalMessage(data.(*hotstufftypes.Proposal))
		}); err != nil {
		conR.Logger.Error("Error adding listener for events", "err", err)
	}
}

func (conR *HotstuffReactor) unsubscribeFromBroadcastEvents() {
	const subscriber = "hotstuff-reactor"
	conR.conS.evsw.RemoveListener(subscriber)
}

func (conR *HotstuffReactor) broadcastProposalMessage(m *hotstufftypes.Proposal) {
	fmt.Println("broadcastProposalMessage", m)
	conR.Switch.Broadcast(p2p.Envelope{
		ChannelID: ProposalChannel,
		Message:   m,
	}, conR.broadcastFunc)
}

func (conR *HotstuffReactor) broadcastNewValidBlockMessage(rs *cstypes.RoundState) {
}

// Broadcasts HasVoteMessage to peers that care.
func (conR *HotstuffReactor) broadcastHasVoteMessage(vote *types.Vote) {
}

func (conR *HotstuffReactor) sendNewRoundStepMessage(peer p2p.Peer) {
}

func (conR *HotstuffReactor) updateRoundStateRoutine() {
}

func (conR *HotstuffReactor) getRoundState() *cstypes.RoundState {
	return &cstypes.RoundState{}
}

func (conR *HotstuffReactor) gossipDataRoutine(peer p2p.Peer, ps *cs.PeerState) {
	// send proposal block parts from proposer to validators
}

func (conR *HotstuffReactor) gossipDataForCatchup(
	logger log.Logger,
	rs *cstypes.RoundState,
	prs *cstypes.PeerRoundState,
	ps *cs.PeerState,
	peer p2p.Peer,
) {
}

func (conR *HotstuffReactor) gossipVotesRoutine(peer p2p.Peer, ps *cs.PeerState) {
}

func (conR *HotstuffReactor) gossipVotesForHeight(
	logger log.Logger,
	rs *cstypes.RoundState,
	prs *cstypes.PeerRoundState,
	ps *cs.PeerState,
) bool {
	return true
}

// NOTE: `queryMaj23Routine` has a simple crude design since it only comes
// into play for liveness when there's a signature DDoS attack happening.
func (conR *HotstuffReactor) queryMaj23Routine(peer p2p.Peer, ps *cs.PeerState) {
}

func (conR *HotstuffReactor) peerStatsRoutine() {
}

func (conR *HotstuffReactor) String() string {
	return "HotstuffConsensusReactor"
}

func (conR *HotstuffReactor) StringIndented(indent string) string {
	return "HotstuffConsensusReactor"
}

//-------------------------------------

// ProposalMessage is sent when a new block is proposed.
type ProposalMessage struct {
	Proposal *types.Proposal
}

// ValidateBasic performs basic validation.
func (m *ProposalMessage) ValidateBasic() error {
	return m.Proposal.ValidateBasic()
}

// String returns a string representation.
func (m *ProposalMessage) String() string {
	return fmt.Sprintf("[Proposal %v]", m.Proposal)
}

//-------------------------------------

// ProposalPOLMessage is sent when a previous proposal is re-proposed.
type ProposalPOLMessage struct {
	Height           int64
	ProposalPOLRound int32
	ProposalPOL      *bits.BitArray
}

// ValidateBasic performs basic validation.
func (m *ProposalPOLMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if m.ProposalPOLRound < 0 {
		return errors.New("negative ProposalPOLRound")
	}
	if m.ProposalPOL.Size() == 0 {
		return errors.New("empty ProposalPOL bit array")
	}
	if m.ProposalPOL.Size() > types.MaxVotesCount {
		return fmt.Errorf("proposalPOL bit array is too big: %d, max: %d", m.ProposalPOL.Size(), types.MaxVotesCount)
	}
	return nil
}

// String returns a string representation.
func (m *ProposalPOLMessage) String() string {
	return fmt.Sprintf("[ProposalPOL H:%v POLR:%v POL:%v]", m.Height, m.ProposalPOLRound, m.ProposalPOL)
}

//-------------------------------------

// BlockPartMessage is sent when gossipping a piece of the proposed block.
type BlockPartMessage struct {
	Height int64
	Round  int32
	Part   *types.Part
}

// ValidateBasic performs basic validation.
func (m *BlockPartMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if m.Round < 0 {
		return errors.New("negative Round")
	}
	if err := m.Part.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong Part: %v", err)
	}
	return nil
}

// String returns a string representation.
func (m *BlockPartMessage) String() string {
	return fmt.Sprintf("[BlockPart H:%v R:%v P:%v]", m.Height, m.Round, m.Part)
}

//-------------------------------------

// VoteMessage is sent when voting for a proposal (or lack thereof).
type VoteMessage struct {
	Vote *types.Vote
}

// ValidateBasic checks whether the vote within the message is well-formed.
func (m *VoteMessage) ValidateBasic() error {
	return m.Vote.ValidateBasic()
}

// String returns a string representation.
func (m *VoteMessage) String() string {
	return fmt.Sprintf("[Vote %v]", m.Vote)
}

//-------------------------------------

// HasVoteMessage is sent to indicate that a particular vote has been received.
type HasVoteMessage struct {
	Height int64
	Round  int32
	Type   cmtproto.SignedMsgType
	Index  int32
}

// ValidateBasic performs basic validation.
func (m *HasVoteMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if m.Round < 0 {
		return errors.New("negative Round")
	}
	if !types.IsVoteTypeValid(m.Type) {
		return errors.New("invalid Type")
	}
	if m.Index < 0 {
		return errors.New("negative Index")
	}
	return nil
}

// String returns a string representation.
func (m *HasVoteMessage) String() string {
	return fmt.Sprintf("[HasVote VI:%v V:{%v/%02d/%v}]", m.Index, m.Height, m.Round, m.Type)
}

//-------------------------------------

// VoteSetMaj23Message is sent to indicate that a given BlockID has seen +2/3 votes.
type VoteSetMaj23Message struct {
	Height  int64
	Round   int32
	Type    cmtproto.SignedMsgType
	BlockID types.BlockID
}

// ValidateBasic performs basic validation.
func (m *VoteSetMaj23Message) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if m.Round < 0 {
		return errors.New("negative Round")
	}
	if !types.IsVoteTypeValid(m.Type) {
		return errors.New("invalid Type")
	}
	if err := m.BlockID.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong BlockID: %v", err)
	}
	return nil
}

// String returns a string representation.
func (m *VoteSetMaj23Message) String() string {
	return fmt.Sprintf("[VSM23 %v/%02d/%v %v]", m.Height, m.Round, m.Type, m.BlockID)
}

//-------------------------------------

// VoteSetBitsMessage is sent to communicate the bit-array of votes seen for the BlockID.
type VoteSetBitsMessage struct {
	Height  int64
	Round   int32
	Type    cmtproto.SignedMsgType
	BlockID types.BlockID
	Votes   *bits.BitArray
}

// ValidateBasic performs basic validation.
func (m *VoteSetBitsMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if !types.IsVoteTypeValid(m.Type) {
		return errors.New("invalid Type")
	}
	if err := m.BlockID.ValidateBasic(); err != nil {
		return fmt.Errorf("wrong BlockID: %v", err)
	}
	// NOTE: Votes.Size() can be zero if the node does not have any
	if m.Votes.Size() > types.MaxVotesCount {
		return fmt.Errorf("votes bit array is too big: %d, max: %d", m.Votes.Size(), types.MaxVotesCount)
	}
	return nil
}

// String returns a string representation.
func (m *VoteSetBitsMessage) String() string {
	return fmt.Sprintf("[VSB %v/%02d/%v %v %v]", m.Height, m.Round, m.Type, m.BlockID, m.Votes)
}

//-------------------------------------

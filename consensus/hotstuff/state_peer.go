package hotstuff

import (
	hotstufftypes "github.com/cometbft/cometbft/consensus/hotstuff/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"sync"
)

type PeerState struct {
	peer   p2p.Peer
	logger log.Logger

	mtx           sync.Mutex
	viewChangeMsg *hotstufftypes.ViewChangeMsg
}

// NewPeerState returns a new PeerState for the given Peer
func NewPeerState(peer p2p.Peer) *PeerState {
	return &PeerState{
		peer:   peer,
		logger: log.NewNopLogger(),
	}
}

// SetLogger allows to set a logger on the peer state. Returns the peer state
// itself.
func (ps *PeerState) SetLogger(logger log.Logger) *PeerState {
	ps.logger = logger
	return ps
}

func (ps *PeerState) setPeerViewChangeMsg(msg *hotstufftypes.ViewChangeMsg) {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()

	if ps.viewChangeMsg != nil && msg.Height < ps.viewChangeMsg.Height {
		return
	}

	if ps.viewChangeMsg != nil && msg.Round <= ps.viewChangeMsg.Round {
		return
	}

	// Check for QC validity
	if ps.viewChangeMsg != nil && msg.HighestKnownQc.Type < ps.viewChangeMsg.HighestKnownQc.Type {
		return
	}

	if !msg.HighestKnownQc.IsValid() {
		return
	}

	ps.viewChangeMsg = msg
}

func (ps *PeerState) getPeerViewChangeMsg() *hotstufftypes.ViewChangeMsg {
	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	return ps.viewChangeMsg
}

package hotstuff

const (
	ProposalChannel = byte(0x80)
	BarChannel      = byte(0x81)
	PingChannel     = byte(0x82)
	PongChannel     = byte(0x83)
	MaxMsgSize      = 1048576 // 1MB
)

type RoundStepType uint8 // These must be numeric, ordered.

// RoundStepType
const (
	RoundStepNewHeight     = RoundStepType(0x01) // Wait til CommitTime + timeoutCommit
	RoundStepNewRound      = RoundStepType(0x02) // Setup new round and go to RoundStepPropose
	RoundStepPropose       = RoundStepType(0x03) // Did propose, gossip proposal
	RoundStepPrevote       = RoundStepType(0x04) // Did prevote, gossip prevotes
	RoundStepPrevoteWait   = RoundStepType(0x05) // Did receive any +2/3 prevotes, start timeout
	RoundStepPrecommit     = RoundStepType(0x06) // Did precommit, gossip precommits
	RoundStepPrecommitWait = RoundStepType(0x07) // Did receive any +2/3 precommits, start timeout
	RoundStepCommit        = RoundStepType(0x08) // Entered commit state machine
	// NOTE: RoundStepNewHeight acts as RoundStepCommitWait.
	// NOTE: Update IsValid method if you change this!
)

const (
	HotstuffPeerStateKey = "HotstuffReactor.peerState"
)

const (
	ProposalEvent = "ProposalEvent"
)

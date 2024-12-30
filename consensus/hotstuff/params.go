package hotstuff

import (
	"fmt"
	"time"
)

const (
	ProposalChannel = byte(0x80)
	BlockChannel    = byte(0x81)
	QCChannel       = byte(0x82)
	VoteChannel     = byte(0x83)
	MaxMsgSize      = 1048576 // 1MB
)

type RoundStepType uint8 // These must be numeric, ordered.

const (
	RoundStepNewView   = RoundStepType(0x01)
	RoundStepPropose   = RoundStepType(0x02)
	RoundStepPrepare   = RoundStepType(0x03)
	RoundStepPreCommit = RoundStepType(0x04)
	RoundStepCommit    = RoundStepType(0x05)
)

type TimeoutInfo struct {
	Duration time.Duration `json:"duration"`
	Height   int64         `json:"height"`
	Round    int32         `json:"round"`
	Step     RoundStepType `json:"step"`
}

func (ti *TimeoutInfo) String() string {
	return fmt.Sprintf("%v ; %d/%d %v", ti.Duration, ti.Height, ti.Round, ti.Step)
}

const (
	HotstuffPeerStateKey = "HotstuffReactor.peerState"
)

const (
	ProposalEvent = "ProposalEvent"
)

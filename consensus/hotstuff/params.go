package hotstuff

import (
	"fmt"
	"time"
)

const (
	ProposalChannel   = byte(0x80)
	VoteChannel       = byte(0x81)
	QCChannel         = byte(0x82)
	ViewChangeChannel = byte(0x83)
	MaxMsgSize        = 1048576 // 1MB
)

type RoundStepType uint8 // These must be numeric, ordered.

const (
	RoundStepViewChange = RoundStepType(0x01)

	RoundStepLeaderPropose   = RoundStepType(0x02)
	RoundStepLeaderPrepare   = RoundStepType(0x03)
	RoundStepLeaderPreCommit = RoundStepType(0x04)
	RoundStepLeaderCommit    = RoundStepType(0x05)

	RoundStepValidatorPropose   = RoundStepType(0x06)
	RoundStepValidatorPrepare   = RoundStepType(0x07)
	RoundStepValidatorPreCommit = RoundStepType(0x08)
	RoundStepValidatorCommit    = RoundStepType(0x09)
)

type TimeoutInfo struct {
	Duration time.Duration `json:"duration"`
	Height   int64         `json:"height"`
	Round    int64         `json:"round"`
	Step     RoundStepType `json:"step"`
}

func (ti *TimeoutInfo) String() string {
	return fmt.Sprintf("%v ; %d/%d %v", ti.Duration, ti.Height, ti.Round, ti.Step)
}

const (
	HotstuffPeerStateKey = "HotstuffReactor.peerState"
)

const (
	ProposalEvent   = "ProposalEvent"
	ViewChangeEvent = "ViewChangeEvent"
	QCEvent         = "QCEvent"
	VoteEvent       = "VoteEvent"
)

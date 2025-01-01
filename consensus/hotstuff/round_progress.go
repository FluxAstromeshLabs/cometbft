package hotstuff

import (
	hotstufftypes "github.com/cometbft/cometbft/consensus/hotstuff/types"
)

type RoundProgress struct {
	Height       int64
	Round        int64
	Proposal     hotstufftypes.Proposal
	PrepareQC    hotstufftypes.QuorumCert
	PreCommitQC  hotstufftypes.QuorumCert
	CommitQC     hotstufftypes.QuorumCert
	ViewChangeQC hotstufftypes.QuorumCert
}

func NewRoundProgress(valSize int, height int64, round int64) *RoundProgress {
	return &RoundProgress{
		Height:   height,
		Round:    round,
		Proposal: hotstufftypes.Proposal{},
		PrepareQC: hotstufftypes.QuorumCert{
			Votes:      make([]byte, valSize),
			Signatures: make([][]byte, valSize),
		},
		PreCommitQC: hotstufftypes.QuorumCert{
			Votes:      make([]byte, valSize),
			Signatures: make([][]byte, valSize),
		},
		CommitQC: hotstufftypes.QuorumCert{
			Votes:      make([]byte, valSize),
			Signatures: make([][]byte, valSize),
		},
		ViewChangeQC: hotstufftypes.QuorumCert{
			Votes:      make([]byte, valSize),
			Signatures: make([][]byte, valSize),
		},
	}
}

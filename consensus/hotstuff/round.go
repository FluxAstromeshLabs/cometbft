package hotstuff

import hotstufftypes "github.com/cometbft/cometbft/consensus/hotstuff/types"

type Round struct {
	Height      int64
	Round       int32
	Proposal    hotstufftypes.Proposal
	PrepareQC   hotstufftypes.QuorumCert
	PreCommitQC hotstufftypes.QuorumCert
	CommitQC    hotstufftypes.QuorumCert
	TimeoutQC   hotstufftypes.QuorumCert
}

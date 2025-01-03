package hotstuff

import (
	"fmt"
	hotstufftypes "github.com/cometbft/cometbft/consensus/hotstuff/types"
)

type RoundProgress struct {
	Height       int64
	Round        int64
	proposal     *hotstufftypes.Proposal
	prepareQC    *hotstufftypes.QuorumCert
	preCommitQC  *hotstufftypes.QuorumCert
	commitQC     *hotstufftypes.QuorumCert
	viewChangeQC *hotstufftypes.QuorumCert
}

func NewRoundProgress(valSize int, height int64, round int64) *RoundProgress {
	return &RoundProgress{
		Height:       height,
		Round:        round,
		proposal:     nil,
		prepareQC:    nil,
		preCommitQC:  nil,
		commitQC:     nil,
		viewChangeQC: nil,
	}
}

func (p *RoundProgress) getProposal() *hotstufftypes.Proposal {
	return p.proposal
}

func (p *RoundProgress) setProposal(proposal *hotstufftypes.Proposal) error {
	validParentQC := true
	if !validParentQC {
		return fmt.Errorf("parent QC is not valid")
	}
	if proposal.Height != p.Height {
		return fmt.Errorf("discarded proposal at height %d for local height %d", proposal.Height, p.Height)
	}
	if proposal.Round != p.Round {
		return fmt.Errorf("discarded proposal at height %d for local height %d", proposal.Height, p.Height)
	}
	p.proposal = proposal
	return nil
}

func (p *RoundProgress) updatePrepareQC(qc *hotstufftypes.QuorumCert) error {
	if qc.Height != p.Height {
		return fmt.Errorf("invalid Prepare QC height %d, expected %d", qc.Height, p.Height)
	}
	if qc.Round != p.Round {
		return fmt.Errorf("invalid Prepare QC round %d, expected %d", qc.Round, p.Round)
	}
	if p.proposal == nil {
		return fmt.Errorf("cannot set prepare QC with empty proposal")
	}
	p.prepareQC = qc
	return nil
}

func (p *RoundProgress) getQC(t hotstufftypes.QCType) *hotstufftypes.QuorumCert {
	switch t {
	case hotstufftypes.PrepareQC:
		return p.prepareQC
	case hotstufftypes.PreCommitQC:
		return p.preCommitQC
	case hotstufftypes.CommitQC:
		return p.commitQC
	case hotstufftypes.ViewChangeQC:
		return p.viewChangeQC
	default:
		return nil
	}
}

package hotstuff

import (
	"fmt"
	hotstufftypes "github.com/cometbft/cometbft/consensus/hotstuff/types"
)

type RoundProgress struct {
	Height         int64
	Round          int64
	highestKnownQC hotstufftypes.QCType

	proposal      *hotstufftypes.Proposal
	viewChangeMsg *hotstufftypes.ViewChangeMsg

	qcs map[hotstufftypes.QCType]*hotstufftypes.QuorumCert
}

func NewRoundProgress(valSize int, height int64, round int64) *RoundProgress {
	return &RoundProgress{
		Height:   height,
		Round:    round,
		proposal: nil,
		qcs:      map[hotstufftypes.QCType]*hotstufftypes.QuorumCert{},
	}
}

func (p *RoundProgress) getProposal() *hotstufftypes.Proposal {
	return p.proposal
}

func (p *RoundProgress) setProposal(proposal *hotstufftypes.Proposal) error {
	if !proposal.Parent_QC.IsValid() {
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

func (p *RoundProgress) getQC(t hotstufftypes.QCType) *hotstufftypes.QuorumCert {
	return p.qcs[t]
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

	p.highestKnownQC = qc.Type
	p.qcs[qc.Type] = qc

	return nil
}

func (p *RoundProgress) setViewChangeMsg(msg *hotstufftypes.ViewChangeMsg) {
	if msg.Height != p.Height {
		return
	}
	if msg.Round > p.Round && msg.HighestKnownQc.Type >= p.highestKnownQC {
		p.Round = msg.Round
		p.highestKnownQC = msg.HighestKnownQc.Type
		p.qcs[msg.HighestKnownQc.Type] = msg.HighestKnownQc
	}
}

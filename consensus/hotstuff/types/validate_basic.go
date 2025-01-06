package types

import "math/bits"

func (m *Proposal) ValidateBasic() error {
	return nil
}

func (m *Vote) ValidateBasic() error {
	return nil
}

func (m *ViewChangeMsg) ValidateBasic() error {
	return nil
}

func (m *QuorumCert) ValidateBasic() error {
	return nil
}

func (m *QuorumCert) IsValid() bool {
	// TODO: validate QC
	return true
}

func (m *QuorumCert) SetVote(signature []byte, idx int32) {
	byteIndex := idx / 8
	bitIndex := idx % 8
	m.Votes[byteIndex] |= 1 << bitIndex
	m.Signatures[idx] = signature
}

func (m *QuorumCert) HasQuorum() bool {
	// TODO: verify validator signatures
	threshold := 2*len(m.Votes)/3 + 1
	voteCount := 0
	for _, b := range m.Votes {
		voteCount += bits.OnesCount8(b)
		if voteCount >= threshold {
			return true
		}
	}
	return voteCount >= threshold
}

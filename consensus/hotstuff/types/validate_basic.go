package types

func (m *Proposal) ValidateBasic() error {
	return nil
}

func (m *Vote) ValidateBasic() error {
	return nil
}

func (m *QuorumCert) ValidateBasic() error {
	return nil
}

func (m *QuorumCert) SetVote(signature []byte, idx int32) {
	byteIndex := idx / 8
	bitIndex := idx % 8
	m.Votes[byteIndex] |= 1 << bitIndex
	m.Signatures[idx] = signature
}

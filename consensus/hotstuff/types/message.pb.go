// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: flux/hotstuff/consensus/message.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type SignedMsgType int32

const (
	UnknownType SignedMsgType = 0
	// Votes
	PrevoteType   SignedMsgType = 1
	PrecommitType SignedMsgType = 2
	// Proposals
	ProposalType SignedMsgType = 32
)

var SignedMsgType_name = map[int32]string{
	0:  "SIGNED_MSG_TYPE_UNKNOWN",
	1:  "SIGNED_MSG_TYPE_PREVOTE",
	2:  "SIGNED_MSG_TYPE_PRECOMMIT",
	32: "SIGNED_MSG_TYPE_PROPOSAL",
}

var SignedMsgType_value = map[string]int32{
	"SIGNED_MSG_TYPE_UNKNOWN":   0,
	"SIGNED_MSG_TYPE_PREVOTE":   1,
	"SIGNED_MSG_TYPE_PRECOMMIT": 2,
	"SIGNED_MSG_TYPE_PROPOSAL":  32,
}

func (x SignedMsgType) String() string {
	return proto.EnumName(SignedMsgType_name, int32(x))
}

func (SignedMsgType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_32138774ec4e69ee, []int{0}
}

type PartSetHeader struct {
	Total uint32 `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	Hash  []byte `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (m *PartSetHeader) Reset()         { *m = PartSetHeader{} }
func (m *PartSetHeader) String() string { return proto.CompactTextString(m) }
func (*PartSetHeader) ProtoMessage()    {}
func (*PartSetHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_32138774ec4e69ee, []int{0}
}
func (m *PartSetHeader) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PartSetHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PartSetHeader.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PartSetHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PartSetHeader.Merge(m, src)
}
func (m *PartSetHeader) XXX_Size() int {
	return m.Size()
}
func (m *PartSetHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_PartSetHeader.DiscardUnknown(m)
}

var xxx_messageInfo_PartSetHeader proto.InternalMessageInfo

func (m *PartSetHeader) GetTotal() uint32 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *PartSetHeader) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

type BlockID struct {
	Hash          []byte        `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	PartSetHeader PartSetHeader `protobuf:"bytes,2,opt,name=part_set_header,json=partSetHeader,proto3" json:"part_set_header"`
}

func (m *BlockID) Reset()         { *m = BlockID{} }
func (m *BlockID) String() string { return proto.CompactTextString(m) }
func (*BlockID) ProtoMessage()    {}
func (*BlockID) Descriptor() ([]byte, []int) {
	return fileDescriptor_32138774ec4e69ee, []int{1}
}
func (m *BlockID) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BlockID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BlockID.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BlockID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockID.Merge(m, src)
}
func (m *BlockID) XXX_Size() int {
	return m.Size()
}
func (m *BlockID) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockID.DiscardUnknown(m)
}

var xxx_messageInfo_BlockID proto.InternalMessageInfo

func (m *BlockID) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *BlockID) GetPartSetHeader() PartSetHeader {
	if m != nil {
		return m.PartSetHeader
	}
	return PartSetHeader{}
}

type Proposal struct {
	Type      SignedMsgType `protobuf:"varint,1,opt,name=type,proto3,enum=flux.hotstuff.consensus.SignedMsgType" json:"type,omitempty"`
	Height    int64         `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
	Round     int32         `protobuf:"varint,3,opt,name=round,proto3" json:"round,omitempty"`
	PolRound  int32         `protobuf:"varint,4,opt,name=pol_round,json=polRound,proto3" json:"pol_round,omitempty"`
	BlockID   BlockID       `protobuf:"bytes,5,opt,name=block_id,json=blockId,proto3" json:"block_id"`
	Timestamp time.Time     `protobuf:"bytes,6,opt,name=timestamp,proto3,stdtime" json:"timestamp"`
	Signature []byte        `protobuf:"bytes,7,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (m *Proposal) Reset()         { *m = Proposal{} }
func (m *Proposal) String() string { return proto.CompactTextString(m) }
func (*Proposal) ProtoMessage()    {}
func (*Proposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_32138774ec4e69ee, []int{2}
}
func (m *Proposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Proposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Proposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Proposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Proposal.Merge(m, src)
}
func (m *Proposal) XXX_Size() int {
	return m.Size()
}
func (m *Proposal) XXX_DiscardUnknown() {
	xxx_messageInfo_Proposal.DiscardUnknown(m)
}

var xxx_messageInfo_Proposal proto.InternalMessageInfo

func (m *Proposal) GetType() SignedMsgType {
	if m != nil {
		return m.Type
	}
	return UnknownType
}

func (m *Proposal) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *Proposal) GetRound() int32 {
	if m != nil {
		return m.Round
	}
	return 0
}

func (m *Proposal) GetPolRound() int32 {
	if m != nil {
		return m.PolRound
	}
	return 0
}

func (m *Proposal) GetBlockID() BlockID {
	if m != nil {
		return m.BlockID
	}
	return BlockID{}
}

func (m *Proposal) GetTimestamp() time.Time {
	if m != nil {
		return m.Timestamp
	}
	return time.Time{}
}

func (m *Proposal) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func init() {
	proto.RegisterEnum("flux.hotstuff.consensus.SignedMsgType", SignedMsgType_name, SignedMsgType_value)
	proto.RegisterType((*PartSetHeader)(nil), "flux.hotstuff.consensus.PartSetHeader")
	proto.RegisterType((*BlockID)(nil), "flux.hotstuff.consensus.BlockID")
	proto.RegisterType((*Proposal)(nil), "flux.hotstuff.consensus.Proposal")
}

func init() {
	proto.RegisterFile("flux/hotstuff/consensus/message.proto", fileDescriptor_32138774ec4e69ee)
}

var fileDescriptor_32138774ec4e69ee = []byte{
	// 585 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x93, 0xcf, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0xbd, 0x6d, 0xda, 0xa6, 0xdb, 0x86, 0x04, 0xab, 0xa2, 0xc6, 0x20, 0xc7, 0x8a, 0x04,
	0x8a, 0x10, 0xb2, 0x11, 0x9c, 0x40, 0xe2, 0x80, 0x69, 0x5a, 0x22, 0xf2, 0xc7, 0x72, 0x5c, 0x2a,
	0xb8, 0x58, 0x4e, 0xbc, 0xb1, 0xad, 0xda, 0x5e, 0xcb, 0xbb, 0x86, 0xf6, 0x0d, 0x50, 0x4e, 0x7d,
	0x81, 0x9c, 0xe0, 0xc0, 0x6b, 0x70, 0xeb, 0xb1, 0x37, 0x38, 0x15, 0x94, 0xbc, 0x08, 0xf2, 0xba,
	0x49, 0xa8, 0x68, 0xb8, 0xed, 0xec, 0xfc, 0x66, 0x3c, 0xf3, 0x7d, 0x5e, 0xf8, 0x60, 0x18, 0xa4,
	0x27, 0xaa, 0x87, 0x29, 0xa1, 0xe9, 0x70, 0xa8, 0x0e, 0x70, 0x44, 0x50, 0x44, 0x52, 0xa2, 0x86,
	0x88, 0x10, 0xdb, 0x45, 0x4a, 0x9c, 0x60, 0x8a, 0xf9, 0xdd, 0x0c, 0x53, 0x66, 0x98, 0x32, 0xc7,
	0xc4, 0x1d, 0x17, 0xbb, 0x98, 0x31, 0x6a, 0x76, 0xca, 0x71, 0xb1, 0xea, 0x62, 0xec, 0x06, 0x48,
	0x65, 0x51, 0x3f, 0x1d, 0xaa, 0xd4, 0x0f, 0x11, 0xa1, 0x76, 0x18, 0xe7, 0x40, 0xed, 0x39, 0x2c,
	0xe9, 0x76, 0x42, 0x7b, 0x88, 0xbe, 0x41, 0xb6, 0x83, 0x12, 0x7e, 0x07, 0xae, 0x51, 0x4c, 0xed,
	0x40, 0x00, 0x32, 0xa8, 0x97, 0x8c, 0x3c, 0xe0, 0x79, 0x58, 0xf0, 0x6c, 0xe2, 0x09, 0x2b, 0x32,
	0xa8, 0x6f, 0x1b, 0xec, 0x5c, 0x23, 0x70, 0x43, 0x0b, 0xf0, 0xe0, 0xb8, 0xb9, 0x37, 0x4f, 0x83,
	0x45, 0x9a, 0x37, 0x61, 0x39, 0xb6, 0x13, 0x6a, 0x11, 0x44, 0x2d, 0x8f, 0xf5, 0x66, 0xd5, 0x5b,
	0x4f, 0x1f, 0x2a, 0x4b, 0x76, 0x50, 0xae, 0x4d, 0xa2, 0x15, 0xce, 0x2f, 0xab, 0x9c, 0x51, 0x8a,
	0xff, 0xbe, 0xac, 0x7d, 0x5f, 0x81, 0x45, 0x3d, 0xc1, 0x31, 0x26, 0x76, 0xc0, 0xbf, 0x80, 0x05,
	0x7a, 0x1a, 0x23, 0xf6, 0xd9, 0x5b, 0xff, 0xe9, 0xdb, 0xf3, 0xdd, 0x08, 0x39, 0x6d, 0xe2, 0x9a,
	0xa7, 0x31, 0x32, 0x58, 0x0d, 0x7f, 0x07, 0xae, 0x7b, 0xc8, 0x77, 0x3d, 0xca, 0xa6, 0x5a, 0x35,
	0xae, 0xa2, 0x6c, 0xff, 0x04, 0xa7, 0x91, 0x23, 0xac, 0xca, 0xa0, 0xbe, 0x66, 0xe4, 0x01, 0x7f,
	0x0f, 0x6e, 0xc6, 0x38, 0xb0, 0xf2, 0x4c, 0x81, 0x65, 0x8a, 0x31, 0x0e, 0x0c, 0x96, 0x6c, 0xc1,
	0x62, 0x3f, 0x13, 0xc2, 0xf2, 0x1d, 0x61, 0x8d, 0xad, 0x28, 0x2f, 0x1d, 0xe5, 0x4a, 0x31, 0xad,
	0x9c, 0x2d, 0x37, 0xb9, 0xac, 0xce, 0x24, 0x34, 0x36, 0x58, 0x8b, 0xa6, 0xc3, 0x6b, 0x70, 0x73,
	0x6e, 0x92, 0xb0, 0xce, 0xda, 0x89, 0x4a, 0x6e, 0xa3, 0x32, 0xb3, 0x51, 0x31, 0x67, 0x84, 0x56,
	0xcc, 0x1a, 0x9d, 0xfd, 0xaa, 0x02, 0x63, 0x51, 0xc6, 0xdf, 0x87, 0x9b, 0xc4, 0x77, 0x23, 0x9b,
	0xa6, 0x09, 0x12, 0x36, 0x98, 0x29, 0x8b, 0x8b, 0x47, 0x3f, 0x00, 0x2c, 0x5d, 0x93, 0x84, 0x7f,
	0x0c, 0x77, 0x7b, 0xcd, 0x83, 0x4e, 0x63, 0xcf, 0x6a, 0xf7, 0x0e, 0x2c, 0xf3, 0xbd, 0xde, 0xb0,
	0x0e, 0x3b, 0x6f, 0x3b, 0xdd, 0xa3, 0x4e, 0x85, 0x13, 0xcb, 0xa3, 0xb1, 0xbc, 0x75, 0x18, 0x1d,
	0x47, 0xf8, 0x53, 0xb4, 0x8c, 0xd6, 0x8d, 0xc6, 0xbb, 0xae, 0xd9, 0xa8, 0x80, 0x9c, 0xd6, 0x13,
	0xf4, 0x11, 0x53, 0xc4, 0xe8, 0x27, 0xf0, 0xee, 0x0d, 0xf4, 0xeb, 0x6e, 0xbb, 0xdd, 0x34, 0x2b,
	0x2b, 0xe2, 0xed, 0xd1, 0x58, 0x2e, 0xe9, 0x09, 0x1a, 0xe0, 0x30, 0xf4, 0x29, 0xab, 0x50, 0xa0,
	0xf0, 0x6f, 0x45, 0x57, 0xef, 0xf6, 0x5e, 0xb5, 0x2a, 0xb2, 0x58, 0x19, 0x8d, 0xe5, 0xed, 0xd9,
	0x2f, 0x90, 0xf1, 0x62, 0xf1, 0xf3, 0x17, 0x89, 0xfb, 0xf6, 0x55, 0x02, 0xda, 0xd1, 0xf9, 0x44,
	0x02, 0x17, 0x13, 0x09, 0xfc, 0x9e, 0x48, 0xe0, 0x6c, 0x2a, 0x71, 0x17, 0x53, 0x89, 0xfb, 0x39,
	0x95, 0xb8, 0x0f, 0x2f, 0x5d, 0x9f, 0x7a, 0x69, 0x5f, 0x19, 0xe0, 0x50, 0xdd, 0x0f, 0xd2, 0x93,
	0xce, 0xbe, 0xd9, 0xb2, 0xfb, 0x44, 0xcd, 0x7c, 0x72, 0xd4, 0x81, 0x67, 0xfb, 0x91, 0x1a, 0x62,
	0x27, 0x0d, 0x10, 0xb9, 0xe1, 0x11, 0xf6, 0xd7, 0x99, 0xf2, 0xcf, 0xfe, 0x04, 0x00, 0x00, 0xff,
	0xff, 0x5e, 0x52, 0x65, 0x76, 0xa6, 0x03, 0x00, 0x00,
}

func (m *PartSetHeader) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PartSetHeader) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PartSetHeader) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Hash) > 0 {
		i -= len(m.Hash)
		copy(dAtA[i:], m.Hash)
		i = encodeVarintMessage(dAtA, i, uint64(len(m.Hash)))
		i--
		dAtA[i] = 0x12
	}
	if m.Total != 0 {
		i = encodeVarintMessage(dAtA, i, uint64(m.Total))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *BlockID) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BlockID) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BlockID) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.PartSetHeader.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintMessage(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Hash) > 0 {
		i -= len(m.Hash)
		copy(dAtA[i:], m.Hash)
		i = encodeVarintMessage(dAtA, i, uint64(len(m.Hash)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Proposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Proposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Proposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Signature) > 0 {
		i -= len(m.Signature)
		copy(dAtA[i:], m.Signature)
		i = encodeVarintMessage(dAtA, i, uint64(len(m.Signature)))
		i--
		dAtA[i] = 0x3a
	}
	n2, err2 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.Timestamp, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.Timestamp):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintMessage(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x32
	{
		size, err := m.BlockID.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintMessage(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	if m.PolRound != 0 {
		i = encodeVarintMessage(dAtA, i, uint64(m.PolRound))
		i--
		dAtA[i] = 0x20
	}
	if m.Round != 0 {
		i = encodeVarintMessage(dAtA, i, uint64(m.Round))
		i--
		dAtA[i] = 0x18
	}
	if m.Height != 0 {
		i = encodeVarintMessage(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x10
	}
	if m.Type != 0 {
		i = encodeVarintMessage(dAtA, i, uint64(m.Type))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintMessage(dAtA []byte, offset int, v uint64) int {
	offset -= sovMessage(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *PartSetHeader) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Total != 0 {
		n += 1 + sovMessage(uint64(m.Total))
	}
	l = len(m.Hash)
	if l > 0 {
		n += 1 + l + sovMessage(uint64(l))
	}
	return n
}

func (m *BlockID) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Hash)
	if l > 0 {
		n += 1 + l + sovMessage(uint64(l))
	}
	l = m.PartSetHeader.Size()
	n += 1 + l + sovMessage(uint64(l))
	return n
}

func (m *Proposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovMessage(uint64(m.Type))
	}
	if m.Height != 0 {
		n += 1 + sovMessage(uint64(m.Height))
	}
	if m.Round != 0 {
		n += 1 + sovMessage(uint64(m.Round))
	}
	if m.PolRound != 0 {
		n += 1 + sovMessage(uint64(m.PolRound))
	}
	l = m.BlockID.Size()
	n += 1 + l + sovMessage(uint64(l))
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.Timestamp)
	n += 1 + l + sovMessage(uint64(l))
	l = len(m.Signature)
	if l > 0 {
		n += 1 + l + sovMessage(uint64(l))
	}
	return n
}

func sovMessage(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMessage(x uint64) (n int) {
	return sovMessage(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PartSetHeader) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PartSetHeader: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PartSetHeader: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Total", wireType)
			}
			m.Total = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Total |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hash", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthMessage
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Hash = append(m.Hash[:0], dAtA[iNdEx:postIndex]...)
			if m.Hash == nil {
				m.Hash = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *BlockID) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BlockID: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BlockID: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hash", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthMessage
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Hash = append(m.Hash[:0], dAtA[iNdEx:postIndex]...)
			if m.Hash == nil {
				m.Hash = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PartSetHeader", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMessage
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.PartSetHeader.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Proposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Proposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Proposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= SignedMsgType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Round", wireType)
			}
			m.Round = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Round |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PolRound", wireType)
			}
			m.PolRound = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PolRound |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockID", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMessage
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.BlockID.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMessage
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.Timestamp, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthMessage
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signature = append(m.Signature[:0], dAtA[iNdEx:postIndex]...)
			if m.Signature == nil {
				m.Signature = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipMessage(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthMessage
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMessage
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMessage
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMessage        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMessage          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMessage = fmt.Errorf("proto: unexpected end of group")
)

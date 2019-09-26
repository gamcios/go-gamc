// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: state.proto

package corepb

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type ConsensusRoot struct {
	Timestamp            int64    `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Proposer             []byte   `protobuf:"bytes,2,opt,name=proposer,proto3" json:"proposer,omitempty"`
	TermRoot             []byte   `protobuf:"bytes,3,opt,name=term_root,json=termRoot,proto3" json:"term_root,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConsensusRoot) Reset()         { *m = ConsensusRoot{} }
func (m *ConsensusRoot) String() string { return proto.CompactTextString(m) }
func (*ConsensusRoot) ProtoMessage()    {}
func (*ConsensusRoot) Descriptor() ([]byte, []int) {
	return fileDescriptor_a888679467bb7853, []int{0}
}
func (m *ConsensusRoot) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusRoot.Unmarshal(m, b)
}
func (m *ConsensusRoot) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusRoot.Marshal(b, m, deterministic)
}
func (m *ConsensusRoot) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusRoot.Merge(m, src)
}
func (m *ConsensusRoot) XXX_Size() int {
	return xxx_messageInfo_ConsensusRoot.Size(m)
}
func (m *ConsensusRoot) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusRoot.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusRoot proto.InternalMessageInfo

func (m *ConsensusRoot) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *ConsensusRoot) GetProposer() []byte {
	if m != nil {
		return m.Proposer
	}
	return nil
}

func (m *ConsensusRoot) GetTermRoot() []byte {
	if m != nil {
		return m.TermRoot
	}
	return nil
}

func init() {
	proto.RegisterType((*ConsensusRoot)(nil), "corepb.ConsensusRoot")
}

func init() { proto.RegisterFile("state.proto", fileDescriptor_a888679467bb7853) }

var fileDescriptor_a888679467bb7853 = []byte{
	// 131 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2e, 0x2e, 0x49, 0x2c,
	0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4b, 0xce, 0x2f, 0x4a, 0x2d, 0x48, 0x52,
	0x4a, 0xe3, 0xe2, 0x75, 0xce, 0xcf, 0x2b, 0x4e, 0xcd, 0x2b, 0x2e, 0x2d, 0x0e, 0xca, 0xcf, 0x2f,
	0x11, 0x92, 0xe1, 0xe2, 0x2c, 0xc9, 0xcc, 0x4d, 0x2d, 0x2e, 0x49, 0xcc, 0x2d, 0x90, 0x60, 0x54,
	0x60, 0xd4, 0x60, 0x0e, 0x42, 0x08, 0x08, 0x49, 0x71, 0x71, 0x14, 0x14, 0xe5, 0x17, 0xe4, 0x17,
	0xa7, 0x16, 0x49, 0x30, 0x29, 0x30, 0x6a, 0xf0, 0x04, 0xc1, 0xf9, 0x42, 0xd2, 0x5c, 0x9c, 0x25,
	0xa9, 0x45, 0xb9, 0xf1, 0x45, 0xf9, 0xf9, 0x25, 0x12, 0xcc, 0x10, 0x49, 0x90, 0x00, 0xc8, 0xd8,
	0x24, 0x36, 0xb0, 0xb5, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x72, 0x09, 0x4d, 0x4d, 0x85,
	0x00, 0x00, 0x00,
}

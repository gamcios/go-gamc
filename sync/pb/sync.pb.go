// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sync.proto

package syncpb

import (
	pb "gamc.pro/gamcio/go-gamc/core/pb"
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

type Sync struct {
	TailBlockHash        []byte   `protobuf:"bytes,1,opt,name=tail_block_hash,json=tailBlockHash,proto3" json:"tail_block_hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Sync) Reset()         { *m = Sync{} }
func (m *Sync) String() string { return proto.CompactTextString(m) }
func (*Sync) ProtoMessage()    {}
func (*Sync) Descriptor() ([]byte, []int) {
	return fileDescriptor_5273b98214de8075, []int{0}
}
func (m *Sync) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Sync.Unmarshal(m, b)
}
func (m *Sync) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Sync.Marshal(b, m, deterministic)
}
func (m *Sync) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Sync.Merge(m, src)
}
func (m *Sync) XXX_Size() int {
	return xxx_messageInfo_Sync.Size(m)
}
func (m *Sync) XXX_DiscardUnknown() {
	xxx_messageInfo_Sync.DiscardUnknown(m)
}

var xxx_messageInfo_Sync proto.InternalMessageInfo

func (m *Sync) GetTailBlockHash() []byte {
	if m != nil {
		return m.TailBlockHash
	}
	return nil
}

type ChunkHeader struct {
	Headers              [][]byte `protobuf:"bytes,1,rep,name=headers,proto3" json:"headers,omitempty"`
	Root                 []byte   `protobuf:"bytes,2,opt,name=root,proto3" json:"root,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChunkHeader) Reset()         { *m = ChunkHeader{} }
func (m *ChunkHeader) String() string { return proto.CompactTextString(m) }
func (*ChunkHeader) ProtoMessage()    {}
func (*ChunkHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_5273b98214de8075, []int{1}
}
func (m *ChunkHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChunkHeader.Unmarshal(m, b)
}
func (m *ChunkHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChunkHeader.Marshal(b, m, deterministic)
}
func (m *ChunkHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChunkHeader.Merge(m, src)
}
func (m *ChunkHeader) XXX_Size() int {
	return xxx_messageInfo_ChunkHeader.Size(m)
}
func (m *ChunkHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_ChunkHeader.DiscardUnknown(m)
}

var xxx_messageInfo_ChunkHeader proto.InternalMessageInfo

func (m *ChunkHeader) GetHeaders() [][]byte {
	if m != nil {
		return m.Headers
	}
	return nil
}

func (m *ChunkHeader) GetRoot() []byte {
	if m != nil {
		return m.Root
	}
	return nil
}

type ChunkHeaders struct {
	ChunkHeaders         []*ChunkHeader `protobuf:"bytes,1,rep,name=chunkHeaders,proto3" json:"chunkHeaders,omitempty"`
	Root                 []byte         `protobuf:"bytes,2,opt,name=root,proto3" json:"root,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *ChunkHeaders) Reset()         { *m = ChunkHeaders{} }
func (m *ChunkHeaders) String() string { return proto.CompactTextString(m) }
func (*ChunkHeaders) ProtoMessage()    {}
func (*ChunkHeaders) Descriptor() ([]byte, []int) {
	return fileDescriptor_5273b98214de8075, []int{2}
}
func (m *ChunkHeaders) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChunkHeaders.Unmarshal(m, b)
}
func (m *ChunkHeaders) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChunkHeaders.Marshal(b, m, deterministic)
}
func (m *ChunkHeaders) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChunkHeaders.Merge(m, src)
}
func (m *ChunkHeaders) XXX_Size() int {
	return xxx_messageInfo_ChunkHeaders.Size(m)
}
func (m *ChunkHeaders) XXX_DiscardUnknown() {
	xxx_messageInfo_ChunkHeaders.DiscardUnknown(m)
}

var xxx_messageInfo_ChunkHeaders proto.InternalMessageInfo

func (m *ChunkHeaders) GetChunkHeaders() []*ChunkHeader {
	if m != nil {
		return m.ChunkHeaders
	}
	return nil
}

func (m *ChunkHeaders) GetRoot() []byte {
	if m != nil {
		return m.Root
	}
	return nil
}

type ChunkData struct {
	Blocks               []*pb.Block `protobuf:"bytes,1,rep,name=blocks,proto3" json:"blocks,omitempty"`
	Root                 []byte      `protobuf:"bytes,2,opt,name=root,proto3" json:"root,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *ChunkData) Reset()         { *m = ChunkData{} }
func (m *ChunkData) String() string { return proto.CompactTextString(m) }
func (*ChunkData) ProtoMessage()    {}
func (*ChunkData) Descriptor() ([]byte, []int) {
	return fileDescriptor_5273b98214de8075, []int{3}
}
func (m *ChunkData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChunkData.Unmarshal(m, b)
}
func (m *ChunkData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChunkData.Marshal(b, m, deterministic)
}
func (m *ChunkData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChunkData.Merge(m, src)
}
func (m *ChunkData) XXX_Size() int {
	return xxx_messageInfo_ChunkData.Size(m)
}
func (m *ChunkData) XXX_DiscardUnknown() {
	xxx_messageInfo_ChunkData.DiscardUnknown(m)
}

var xxx_messageInfo_ChunkData proto.InternalMessageInfo

func (m *ChunkData) GetBlocks() []*pb.Block {
	if m != nil {
		return m.Blocks
	}
	return nil
}

func (m *ChunkData) GetRoot() []byte {
	if m != nil {
		return m.Root
	}
	return nil
}

func init() {
	proto.RegisterType((*Sync)(nil), "syncpb.Sync")
	proto.RegisterType((*ChunkHeader)(nil), "syncpb.ChunkHeader")
	proto.RegisterType((*ChunkHeaders)(nil), "syncpb.ChunkHeaders")
	proto.RegisterType((*ChunkData)(nil), "syncpb.ChunkData")
}

func init() { proto.RegisterFile("sync.proto", fileDescriptor_5273b98214de8075) }

var fileDescriptor_5273b98214de8075 = []byte{
	// 228 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0x41, 0x4b, 0x03, 0x31,
	0x10, 0x85, 0x59, 0x2d, 0x2b, 0x4e, 0x53, 0x84, 0x78, 0x09, 0x9e, 0xca, 0x82, 0x52, 0x10, 0x13,
	0xd0, 0x83, 0x07, 0x6f, 0x2a, 0xd2, 0x73, 0x3c, 0x7a, 0x28, 0x49, 0x5c, 0x4c, 0x69, 0xd9, 0x09,
	0x49, 0x3c, 0xec, 0xbf, 0x97, 0xcc, 0xba, 0xb0, 0x85, 0xbd, 0xbd, 0x79, 0x33, 0xef, 0x83, 0x37,
	0x00, 0xa9, 0xef, 0x9c, 0x0c, 0x11, 0x33, 0xf2, 0xba, 0xe8, 0x60, 0x6f, 0xee, 0x5d, 0x32, 0xe4,
	0xa9, 0x22, 0xf6, 0xa8, 0x7e, 0xf0, 0xa1, 0x28, 0xe5, 0x30, 0xb6, 0x2a, 0x58, 0x65, 0x8f, 0xe8,
	0x0e, 0x43, 0xa8, 0x91, 0xb0, 0xf8, 0xec, 0x3b, 0xc7, 0xef, 0xe0, 0x2a, 0x9b, 0xfd, 0x71, 0x47,
	0xbb, 0x9d, 0x37, 0xc9, 0x8b, 0x6a, 0x5d, 0x6d, 0x98, 0x5e, 0x15, 0xfb, 0xb5, 0xb8, 0x5b, 0x93,
	0x7c, 0xf3, 0x02, 0xcb, 0x37, 0xff, 0xdb, 0x1d, 0xb6, 0xad, 0xf9, 0x6e, 0x23, 0x17, 0x70, 0xe1,
	0x49, 0x25, 0x51, 0xad, 0xcf, 0x37, 0x4c, 0x8f, 0x23, 0xe7, 0xb0, 0x88, 0x88, 0x59, 0x9c, 0x11,
	0x85, 0x74, 0xf3, 0x05, 0x6c, 0x12, 0x4e, 0xfc, 0x19, 0x98, 0x9b, 0xcc, 0x84, 0x58, 0x3e, 0x5e,
	0xcb, 0xa1, 0x88, 0x9c, 0xdc, 0xea, 0x93, 0xc3, 0x59, 0xf8, 0x07, 0x5c, 0x52, 0xe0, 0xdd, 0x64,
	0xc3, 0x6f, 0xa1, 0xa6, 0x26, 0x23, 0x73, 0x25, 0x4b, 0xf9, 0x60, 0x25, 0x35, 0xd1, 0xff, 0xcb,
	0x39, 0x8e, 0xad, 0xe9, 0x31, 0x4f, 0x7f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x6d, 0xc6, 0x97, 0xbc,
	0x5b, 0x01, 0x00, 0x00,
}

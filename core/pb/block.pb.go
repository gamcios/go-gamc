// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: block.proto

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

type Data struct {
	Type                 string   `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Msg                  []byte   `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Data) Reset()         { *m = Data{} }
func (m *Data) String() string { return proto.CompactTextString(m) }
func (*Data) ProtoMessage()    {}
func (*Data) Descriptor() ([]byte, []int) {
	return fileDescriptor_8e550b1f5926e92d, []int{0}
}
func (m *Data) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data.Unmarshal(m, b)
}
func (m *Data) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data.Marshal(b, m, deterministic)
}
func (m *Data) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data.Merge(m, src)
}
func (m *Data) XXX_Size() int {
	return xxx_messageInfo_Data.Size(m)
}
func (m *Data) XXX_DiscardUnknown() {
	xxx_messageInfo_Data.DiscardUnknown(m)
}

var xxx_messageInfo_Data proto.InternalMessageInfo

func (m *Data) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Data) GetMsg() []byte {
	if m != nil {
		return m.Msg
	}
	return nil
}

type Signature struct {
	Signer               []byte   `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	Data                 []byte   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Signature) Reset()         { *m = Signature{} }
func (m *Signature) String() string { return proto.CompactTextString(m) }
func (*Signature) ProtoMessage()    {}
func (*Signature) Descriptor() ([]byte, []int) {
	return fileDescriptor_8e550b1f5926e92d, []int{1}
}
func (m *Signature) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Signature.Unmarshal(m, b)
}
func (m *Signature) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Signature.Marshal(b, m, deterministic)
}
func (m *Signature) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Signature.Merge(m, src)
}
func (m *Signature) XXX_Size() int {
	return xxx_messageInfo_Signature.Size(m)
}
func (m *Signature) XXX_DiscardUnknown() {
	xxx_messageInfo_Signature.DiscardUnknown(m)
}

var xxx_messageInfo_Signature proto.InternalMessageInfo

func (m *Signature) GetSigner() []byte {
	if m != nil {
		return m.Signer
	}
	return nil
}

func (m *Signature) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type Transaction struct {
	Hash                 []byte     `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	From                 []byte     `protobuf:"bytes,2,opt,name=from,proto3" json:"from,omitempty"`
	To                   []byte     `protobuf:"bytes,3,opt,name=to,proto3" json:"to,omitempty"`
	Value                []byte     `protobuf:"bytes,4,opt,name=value,proto3" json:"value,omitempty"`
	Nonce                uint64     `protobuf:"varint,5,opt,name=nonce,proto3" json:"nonce,omitempty"`
	ChainId              uint32     `protobuf:"varint,6,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	Fee                  []byte     `protobuf:"bytes,7,opt,name=fee,proto3" json:"fee,omitempty"`
	Timestamp            int64      `protobuf:"varint,8,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Data                 *Data      `protobuf:"bytes,9,opt,name=data,proto3" json:"data,omitempty"`
	Priority             uint32     `protobuf:"varint,10,opt,name=priority,proto3" json:"priority,omitempty"`
	Sign                 *Signature `protobuf:"bytes,11,opt,name=sign,proto3" json:"sign,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Transaction) Reset()         { *m = Transaction{} }
func (m *Transaction) String() string { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()    {}
func (*Transaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_8e550b1f5926e92d, []int{2}
}
func (m *Transaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction.Unmarshal(m, b)
}
func (m *Transaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction.Marshal(b, m, deterministic)
}
func (m *Transaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction.Merge(m, src)
}
func (m *Transaction) XXX_Size() int {
	return xxx_messageInfo_Transaction.Size(m)
}
func (m *Transaction) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction proto.InternalMessageInfo

func (m *Transaction) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *Transaction) GetFrom() []byte {
	if m != nil {
		return m.From
	}
	return nil
}

func (m *Transaction) GetTo() []byte {
	if m != nil {
		return m.To
	}
	return nil
}

func (m *Transaction) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *Transaction) GetNonce() uint64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func (m *Transaction) GetChainId() uint32 {
	if m != nil {
		return m.ChainId
	}
	return 0
}

func (m *Transaction) GetFee() []byte {
	if m != nil {
		return m.Fee
	}
	return nil
}

func (m *Transaction) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *Transaction) GetData() *Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Transaction) GetPriority() uint32 {
	if m != nil {
		return m.Priority
	}
	return 0
}

func (m *Transaction) GetSign() *Signature {
	if m != nil {
		return m.Sign
	}
	return nil
}

type Witness struct {
	Master               []byte   `protobuf:"bytes,1,opt,name=master,proto3" json:"master,omitempty"`
	Followers            [][]byte `protobuf:"bytes,2,rep,name=followers,proto3" json:"followers,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Witness) Reset()         { *m = Witness{} }
func (m *Witness) String() string { return proto.CompactTextString(m) }
func (*Witness) ProtoMessage()    {}
func (*Witness) Descriptor() ([]byte, []int) {
	return fileDescriptor_8e550b1f5926e92d, []int{3}
}
func (m *Witness) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Witness.Unmarshal(m, b)
}
func (m *Witness) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Witness.Marshal(b, m, deterministic)
}
func (m *Witness) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Witness.Merge(m, src)
}
func (m *Witness) XXX_Size() int {
	return xxx_messageInfo_Witness.Size(m)
}
func (m *Witness) XXX_DiscardUnknown() {
	xxx_messageInfo_Witness.DiscardUnknown(m)
}

var xxx_messageInfo_Witness proto.InternalMessageInfo

func (m *Witness) GetMaster() []byte {
	if m != nil {
		return m.Master
	}
	return nil
}

func (m *Witness) GetFollowers() [][]byte {
	if m != nil {
		return m.Followers
	}
	return nil
}

type PsecData struct {
	Term                 int64    `protobuf:"varint,1,opt,name=term,proto3" json:"term,omitempty"`
	Timestamp            int64    `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PsecData) Reset()         { *m = PsecData{} }
func (m *PsecData) String() string { return proto.CompactTextString(m) }
func (*PsecData) ProtoMessage()    {}
func (*PsecData) Descriptor() ([]byte, []int) {
	return fileDescriptor_8e550b1f5926e92d, []int{4}
}
func (m *PsecData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PsecData.Unmarshal(m, b)
}
func (m *PsecData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PsecData.Marshal(b, m, deterministic)
}
func (m *PsecData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PsecData.Merge(m, src)
}
func (m *PsecData) XXX_Size() int {
	return xxx_messageInfo_PsecData.Size(m)
}
func (m *PsecData) XXX_DiscardUnknown() {
	xxx_messageInfo_PsecData.DiscardUnknown(m)
}

var xxx_messageInfo_PsecData proto.InternalMessageInfo

func (m *PsecData) GetTerm() int64 {
	if m != nil {
		return m.Term
	}
	return 0
}

func (m *PsecData) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

type BlockHeader struct {
	Hash                 []byte     `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	ParentHash           []byte     `protobuf:"bytes,2,opt,name=parent_hash,json=parentHash,proto3" json:"parent_hash,omitempty"`
	Coinbase             []byte     `protobuf:"bytes,3,opt,name=coinbase,proto3" json:"coinbase,omitempty"`
	Timestamp            int64      `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	ChainId              uint32     `protobuf:"varint,5,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	Height               uint64     `protobuf:"varint,6,opt,name=height,proto3" json:"height,omitempty"`
	WitnessReward        []byte     `protobuf:"bytes,7,opt,name=witness_reward,json=witnessReward,proto3" json:"witness_reward,omitempty"`
	Witnesses            []*Witness `protobuf:"bytes,8,rep,name=witnesses,proto3" json:"witnesses,omitempty"`
	StateRoot            []byte     `protobuf:"bytes,9,opt,name=state_root,json=stateRoot,proto3" json:"state_root,omitempty"`
	TxsRoot              []byte     `protobuf:"bytes,10,opt,name=txs_root,json=txsRoot,proto3" json:"txs_root,omitempty"`
	PsecData             *PsecData  `protobuf:"bytes,11,opt,name=psec_data,json=psecData,proto3" json:"psec_data,omitempty"`
	Sign                 *Signature `protobuf:"bytes,12,opt,name=sign,proto3" json:"sign,omitempty"`
	Extra                []byte     `protobuf:"bytes,13,opt,name=extra,proto3" json:"extra,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *BlockHeader) Reset()         { *m = BlockHeader{} }
func (m *BlockHeader) String() string { return proto.CompactTextString(m) }
func (*BlockHeader) ProtoMessage()    {}
func (*BlockHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_8e550b1f5926e92d, []int{5}
}
func (m *BlockHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockHeader.Unmarshal(m, b)
}
func (m *BlockHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockHeader.Marshal(b, m, deterministic)
}
func (m *BlockHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockHeader.Merge(m, src)
}
func (m *BlockHeader) XXX_Size() int {
	return xxx_messageInfo_BlockHeader.Size(m)
}
func (m *BlockHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockHeader.DiscardUnknown(m)
}

var xxx_messageInfo_BlockHeader proto.InternalMessageInfo

func (m *BlockHeader) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *BlockHeader) GetParentHash() []byte {
	if m != nil {
		return m.ParentHash
	}
	return nil
}

func (m *BlockHeader) GetCoinbase() []byte {
	if m != nil {
		return m.Coinbase
	}
	return nil
}

func (m *BlockHeader) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *BlockHeader) GetChainId() uint32 {
	if m != nil {
		return m.ChainId
	}
	return 0
}

func (m *BlockHeader) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *BlockHeader) GetWitnessReward() []byte {
	if m != nil {
		return m.WitnessReward
	}
	return nil
}

func (m *BlockHeader) GetWitnesses() []*Witness {
	if m != nil {
		return m.Witnesses
	}
	return nil
}

func (m *BlockHeader) GetStateRoot() []byte {
	if m != nil {
		return m.StateRoot
	}
	return nil
}

func (m *BlockHeader) GetTxsRoot() []byte {
	if m != nil {
		return m.TxsRoot
	}
	return nil
}

func (m *BlockHeader) GetPsecData() *PsecData {
	if m != nil {
		return m.PsecData
	}
	return nil
}

func (m *BlockHeader) GetSign() *Signature {
	if m != nil {
		return m.Sign
	}
	return nil
}

func (m *BlockHeader) GetExtra() []byte {
	if m != nil {
		return m.Extra
	}
	return nil
}

type Block struct {
	Hash                 []byte         `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Header               *BlockHeader   `protobuf:"bytes,2,opt,name=header,proto3" json:"header,omitempty"`
	Body                 []*Transaction `protobuf:"bytes,3,rep,name=body,proto3" json:"body,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Block) Reset()         { *m = Block{} }
func (m *Block) String() string { return proto.CompactTextString(m) }
func (*Block) ProtoMessage()    {}
func (*Block) Descriptor() ([]byte, []int) {
	return fileDescriptor_8e550b1f5926e92d, []int{6}
}
func (m *Block) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Block.Unmarshal(m, b)
}
func (m *Block) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Block.Marshal(b, m, deterministic)
}
func (m *Block) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Block.Merge(m, src)
}
func (m *Block) XXX_Size() int {
	return xxx_messageInfo_Block.Size(m)
}
func (m *Block) XXX_DiscardUnknown() {
	xxx_messageInfo_Block.DiscardUnknown(m)
}

var xxx_messageInfo_Block proto.InternalMessageInfo

func (m *Block) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *Block) GetHeader() *BlockHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *Block) GetBody() []*Transaction {
	if m != nil {
		return m.Body
	}
	return nil
}

type DownloadBlock struct {
	Hash                 []byte     `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Sign                 *Signature `protobuf:"bytes,2,opt,name=sign,proto3" json:"sign,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *DownloadBlock) Reset()         { *m = DownloadBlock{} }
func (m *DownloadBlock) String() string { return proto.CompactTextString(m) }
func (*DownloadBlock) ProtoMessage()    {}
func (*DownloadBlock) Descriptor() ([]byte, []int) {
	return fileDescriptor_8e550b1f5926e92d, []int{7}
}
func (m *DownloadBlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DownloadBlock.Unmarshal(m, b)
}
func (m *DownloadBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DownloadBlock.Marshal(b, m, deterministic)
}
func (m *DownloadBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DownloadBlock.Merge(m, src)
}
func (m *DownloadBlock) XXX_Size() int {
	return xxx_messageInfo_DownloadBlock.Size(m)
}
func (m *DownloadBlock) XXX_DiscardUnknown() {
	xxx_messageInfo_DownloadBlock.DiscardUnknown(m)
}

var xxx_messageInfo_DownloadBlock proto.InternalMessageInfo

func (m *DownloadBlock) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *DownloadBlock) GetSign() *Signature {
	if m != nil {
		return m.Sign
	}
	return nil
}

func init() {
	proto.RegisterType((*Data)(nil), "corepb.Data")
	proto.RegisterType((*Signature)(nil), "corepb.Signature")
	proto.RegisterType((*Transaction)(nil), "corepb.Transaction")
	proto.RegisterType((*Witness)(nil), "corepb.Witness")
	proto.RegisterType((*PsecData)(nil), "corepb.PsecData")
	proto.RegisterType((*BlockHeader)(nil), "corepb.BlockHeader")
	proto.RegisterType((*Block)(nil), "corepb.Block")
	proto.RegisterType((*DownloadBlock)(nil), "corepb.DownloadBlock")
}

func init() { proto.RegisterFile("block.proto", fileDescriptor_8e550b1f5926e92d) }

var fileDescriptor_8e550b1f5926e92d = []byte{
	// 588 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x94, 0x4d, 0x6f, 0xd3, 0x4c,
	0x10, 0xc7, 0xe5, 0x97, 0x24, 0xf6, 0xd8, 0xe9, 0xd3, 0x67, 0x41, 0xd5, 0x52, 0x81, 0xb0, 0x2c,
	0x55, 0x58, 0x82, 0xf6, 0x50, 0x0e, 0x5c, 0x90, 0x90, 0x50, 0x0f, 0x85, 0x13, 0x5a, 0x90, 0x38,
	0x46, 0x1b, 0x7b, 0x93, 0x58, 0xc4, 0x5e, 0xb3, 0xbb, 0x25, 0xed, 0xc7, 0xe1, 0xcc, 0x97, 0x44,
	0x3b, 0x6b, 0xe7, 0xa5, 0x40, 0x6f, 0x33, 0xff, 0x9d, 0xf1, 0xee, 0xfc, 0x7f, 0x93, 0x40, 0x32,
	0x5f, 0xcb, 0xf2, 0xdb, 0x45, 0xa7, 0xa4, 0x91, 0x64, 0x5c, 0x4a, 0x25, 0xba, 0x79, 0xfe, 0x0a,
	0xc2, 0x2b, 0x6e, 0x38, 0x21, 0x10, 0x9a, 0xbb, 0x4e, 0x50, 0x2f, 0xf3, 0x8a, 0x98, 0x61, 0x4c,
	0x8e, 0x21, 0x68, 0xf4, 0x92, 0xfa, 0x99, 0x57, 0xa4, 0xcc, 0x86, 0xf9, 0x1b, 0x88, 0x3f, 0xd7,
	0xcb, 0x96, 0x9b, 0x1b, 0x25, 0xc8, 0x09, 0x8c, 0x75, 0xbd, 0x6c, 0x85, 0xc2, 0xa6, 0x94, 0xf5,
	0x99, 0xfd, 0x54, 0xc5, 0x0d, 0xef, 0xfb, 0x30, 0xce, 0x7f, 0xfa, 0x90, 0x7c, 0x51, 0xbc, 0xd5,
	0xbc, 0x34, 0xb5, 0x6c, 0x6d, 0xcd, 0x8a, 0xeb, 0x55, 0xdf, 0x89, 0xb1, 0xd5, 0x16, 0x4a, 0x36,
	0x43, 0x9f, 0x8d, 0xc9, 0x11, 0xf8, 0x46, 0xd2, 0x00, 0x15, 0xdf, 0x48, 0xf2, 0x18, 0x46, 0x3f,
	0xf8, 0xfa, 0x46, 0xd0, 0x10, 0x25, 0x97, 0x58, 0xb5, 0x95, 0x6d, 0x29, 0xe8, 0x28, 0xf3, 0x8a,
	0x90, 0xb9, 0x84, 0x3c, 0x81, 0xa8, 0x5c, 0xf1, 0xba, 0x9d, 0xd5, 0x15, 0x1d, 0x67, 0x5e, 0x31,
	0x65, 0x13, 0xcc, 0x3f, 0x54, 0x76, 0xb2, 0x85, 0x10, 0x74, 0xe2, 0x26, 0x5b, 0x08, 0x41, 0x9e,
	0x42, 0x6c, 0xea, 0x46, 0x68, 0xc3, 0x9b, 0x8e, 0x46, 0x99, 0x57, 0x04, 0x6c, 0x27, 0x90, 0xac,
	0x1f, 0x29, 0xce, 0xbc, 0x22, 0xb9, 0x4c, 0x2f, 0x9c, 0x79, 0x17, 0xd6, 0x39, 0x37, 0x20, 0x39,
	0x85, 0xa8, 0x53, 0xb5, 0x54, 0xb5, 0xb9, 0xa3, 0x80, 0x97, 0x6d, 0x73, 0x72, 0x06, 0xa1, 0xb5,
	0x86, 0x26, 0xd8, 0xfd, 0xff, 0xd0, 0xbd, 0x75, 0x92, 0xe1, 0x71, 0xfe, 0x0e, 0x26, 0x5f, 0x6b,
	0xd3, 0x0a, 0xad, 0xad, 0xb5, 0x0d, 0xd7, 0x66, 0x67, 0xad, 0xcb, 0xec, 0x2b, 0x17, 0x72, 0xbd,
	0x96, 0x1b, 0xa1, 0x34, 0xf5, 0xb3, 0xa0, 0x48, 0xd9, 0x4e, 0xc8, 0xdf, 0x42, 0xf4, 0x49, 0x8b,
	0x72, 0xcb, 0x53, 0xa8, 0x06, 0xfb, 0x03, 0x86, 0xf1, 0xe1, 0x8c, 0xfe, 0xbd, 0x19, 0xf3, 0x5f,
	0x01, 0x24, 0xef, 0xed, 0x86, 0x5c, 0x0b, 0x5e, 0x39, 0x8c, 0x7f, 0x20, 0x7a, 0x0e, 0x49, 0xc7,
	0x95, 0x68, 0xcd, 0x0c, 0x8f, 0x1c, 0x29, 0x70, 0xd2, 0xb5, 0x2d, 0x38, 0x85, 0xa8, 0x94, 0x75,
	0x3b, 0xe7, 0x5a, 0xf4, 0xd4, 0xb6, 0xf9, 0xe1, 0xf5, 0xe1, 0x7d, 0x8b, 0xf7, 0x69, 0x8d, 0x0e,
	0x69, 0x9d, 0xc0, 0x78, 0x25, 0xea, 0xe5, 0xca, 0x20, 0xc6, 0x90, 0xf5, 0x19, 0x39, 0x83, 0xa3,
	0x8d, 0x33, 0x6c, 0xa6, 0xc4, 0x86, 0xab, 0xaa, 0x07, 0x3a, 0xed, 0x55, 0x86, 0x22, 0x39, 0x87,
	0xb8, 0x17, 0x84, 0xa6, 0x51, 0x16, 0x14, 0xc9, 0xe5, 0x7f, 0x03, 0x83, 0xde, 0x70, 0xb6, 0xab,
	0x20, 0xcf, 0x00, 0xb4, 0xe1, 0x46, 0xcc, 0x94, 0x94, 0x06, 0x89, 0xa7, 0x2c, 0x46, 0x85, 0x49,
	0x69, 0xec, 0x3b, 0xcd, 0xad, 0x76, 0x87, 0x80, 0x87, 0x13, 0x73, 0xab, 0xf1, 0xe8, 0x1c, 0xe2,
	0x4e, 0x8b, 0x72, 0x86, 0xab, 0xe2, 0x60, 0x1f, 0x0f, 0x17, 0x0d, 0x60, 0x58, 0xd4, 0x0d, 0x88,
	0x86, 0xb5, 0x48, 0x1f, 0x5c, 0x0b, 0xbb, 0xdc, 0xe2, 0xd6, 0x28, 0x4e, 0xa7, 0x6e, 0xe5, 0x31,
	0xc9, 0xbf, 0xc3, 0x08, 0x61, 0xfd, 0x15, 0xd3, 0x4b, 0x6b, 0x98, 0x85, 0x88, 0x84, 0x92, 0xcb,
	0x47, 0xc3, 0xb7, 0xf7, 0xf8, 0xb2, 0xbe, 0x84, 0xbc, 0x80, 0x70, 0x2e, 0xab, 0x3b, 0x1a, 0xa0,
	0x33, 0xdb, 0xd2, 0xbd, 0x5f, 0x2b, 0xc3, 0x82, 0xfc, 0x23, 0x4c, 0xaf, 0xe4, 0xa6, 0x5d, 0x4b,
	0x5e, 0xfd, 0xfb, 0xea, 0x61, 0x28, 0xff, 0xc1, 0xa1, 0xe6, 0x63, 0xfc, 0x17, 0x7a, 0xfd, 0x3b,
	0x00, 0x00, 0xff, 0xff, 0xf5, 0x54, 0xc0, 0xe2, 0x94, 0x04, 0x00, 0x00,
}
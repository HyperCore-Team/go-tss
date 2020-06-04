// Code generated by protoc-gen-go. DO NOT EDIT.
// source: signature_notifier.proto

package messages

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = proto.Marshal
	_ = fmt.Errorf
	_ = math.Inf
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type KeysignSignature_Status int32

const (
	KeysignSignature_Unknown KeysignSignature_Status = 0
	KeysignSignature_Success KeysignSignature_Status = 1
	KeysignSignature_Failed  KeysignSignature_Status = 2
)

var KeysignSignature_Status_name = map[int32]string{
	0: "Unknown",
	1: "Success",
	2: "Failed",
}

var KeysignSignature_Status_value = map[string]int32{
	"Unknown": 0,
	"Success": 1,
	"Failed":  2,
}

func (x KeysignSignature_Status) String() string {
	return proto.EnumName(KeysignSignature_Status_name, int32(x))
}

func (KeysignSignature_Status) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_7604b65b7d1ea1e3, []int{0, 0}
}

type KeysignSignature struct {
	ID                   string                  `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Signature            []byte                  `protobuf:"bytes,2,opt,name=Signature,proto3" json:"Signature,omitempty"`
	KeysignStatus        KeysignSignature_Status `protobuf:"varint,3,opt,name=KeysignStatus,proto3,enum=messages.KeysignSignature_Status" json:"KeysignStatus,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *KeysignSignature) Reset()         { *m = KeysignSignature{} }
func (m *KeysignSignature) String() string { return proto.CompactTextString(m) }
func (*KeysignSignature) ProtoMessage()    {}
func (*KeysignSignature) Descriptor() ([]byte, []int) {
	return fileDescriptor_7604b65b7d1ea1e3, []int{0}
}

func (m *KeysignSignature) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeysignSignature.Unmarshal(m, b)
}

func (m *KeysignSignature) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeysignSignature.Marshal(b, m, deterministic)
}

func (m *KeysignSignature) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeysignSignature.Merge(m, src)
}

func (m *KeysignSignature) XXX_Size() int {
	return xxx_messageInfo_KeysignSignature.Size(m)
}

func (m *KeysignSignature) XXX_DiscardUnknown() {
	xxx_messageInfo_KeysignSignature.DiscardUnknown(m)
}

var xxx_messageInfo_KeysignSignature proto.InternalMessageInfo

func (m *KeysignSignature) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *KeysignSignature) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *KeysignSignature) GetKeysignStatus() KeysignSignature_Status {
	if m != nil {
		return m.KeysignStatus
	}
	return KeysignSignature_Unknown
}

func init() {
	proto.RegisterEnum("messages.KeysignSignature_Status", KeysignSignature_Status_name, KeysignSignature_Status_value)
	proto.RegisterType((*KeysignSignature)(nil), "messages.KeysignSignature")
}

func init() { proto.RegisterFile("signature_notifier.proto", fileDescriptor_7604b65b7d1ea1e3) }

var fileDescriptor_7604b65b7d1ea1e3 = []byte{
	// 182 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x28, 0xce, 0x4c, 0xcf,
	0x4b, 0x2c, 0x29, 0x2d, 0x4a, 0x8d, 0xcf, 0xcb, 0x2f, 0xc9, 0x4c, 0xcb, 0x4c, 0x2d, 0xd2, 0x2b,
	0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0xc8, 0x4d, 0x2d, 0x2e, 0x4e, 0x4c, 0x4f, 0x2d, 0x56, 0xda,
	0xc9, 0xc8, 0x25, 0xe0, 0x9d, 0x5a, 0x09, 0x52, 0x19, 0x0c, 0x53, 0x2d, 0xc4, 0xc7, 0xc5, 0xe4,
	0xe9, 0x22, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0x19, 0xc4, 0xe4, 0xe9, 0x22, 0x24, 0xc3, 0xc5, 0x09,
	0x97, 0x94, 0x60, 0x52, 0x60, 0xd4, 0xe0, 0x09, 0x42, 0x08, 0x08, 0xb9, 0x73, 0xf1, 0xc2, 0x4c,
	0x28, 0x49, 0x2c, 0x29, 0x2d, 0x96, 0x60, 0x56, 0x60, 0xd4, 0xe0, 0x33, 0x52, 0xd4, 0x83, 0x59,
	0xa2, 0x87, 0x6e, 0x81, 0x1e, 0x44, 0x61, 0x10, 0xaa, 0x3e, 0x25, 0x3d, 0x2e, 0x36, 0x08, 0x4b,
	0x88, 0x9b, 0x8b, 0x3d, 0x34, 0x2f, 0x3b, 0x2f, 0xbf, 0x3c, 0x4f, 0x80, 0x01, 0xc4, 0x09, 0x2e,
	0x4d, 0x4e, 0x4e, 0x2d, 0x2e, 0x16, 0x60, 0x14, 0xe2, 0xe2, 0x62, 0x73, 0x4b, 0xcc, 0xcc, 0x49,
	0x4d, 0x11, 0x60, 0x4a, 0x62, 0x03, 0x7b, 0xc6, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0xfa, 0x06,
	0x4f, 0xd8, 0xe8, 0x00, 0x00, 0x00,
}

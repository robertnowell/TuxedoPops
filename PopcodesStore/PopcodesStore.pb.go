// Code generated by protoc-gen-go.
// source: PopcodesStore.proto
// DO NOT EDIT!

/*
Package PopcodesStore is a generated protocol buffer package.

It is generated from these files:
	PopcodesStore.proto

It has these top-level messages:
	Popcodes
	OTX
*/
package PopcodesStore

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Popcodes struct {
	Address string `protobuf:"bytes,1,opt,name=Address" json:"Address,omitempty"`
	Counter []byte `protobuf:"bytes,3,opt,name=Counter,proto3" json:"Counter,omitempty"`
	Outputs []*OTX `protobuf:"bytes,4,rep,name=Outputs" json:"Outputs,omitempty"`
}

func (m *Popcodes) Reset()         { *m = Popcodes{} }
func (m *Popcodes) String() string { return proto.CompactTextString(m) }
func (*Popcodes) ProtoMessage()    {}

func (m *Popcodes) GetOutputs() []*OTX {
	if m != nil {
		return m.Outputs
	}
	return nil
}

type OTX struct {
	Owners    [][]byte `protobuf:"bytes,1,rep,name=Owners,proto3" json:"Owners,omitempty"`
	Threshold int64    `protobuf:"varint,2,opt,name=Threshold" json:"Threshold,omitempty"`
	Amount    int64    `protobuf:"varint,3,opt,name=Amount" json:"Amount,omitempty"`
	Data      string   `protobuf:"bytes,4,opt,name=Data" json:"Data,omitempty"`
	Creator   []byte   `protobuf:"bytes,5,opt,name=Creator,proto3" json:"Creator,omitempty"`
}

func (m *OTX) Reset()         { *m = OTX{} }
func (m *OTX) String() string { return proto.CompactTextString(m) }
func (*OTX) ProtoMessage()    {}

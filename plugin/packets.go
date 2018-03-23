package plugin

import (
	"fmt"
)

// TODO: refactor to use capital (exported) types?
type sstpHeader struct {
	MajorVersion uint8
	MinorVersion uint8
	C            bool
	Length       uint16
}

// MessageType is the type of message this packet is
type MessageType uint16

// Constants for MessageType values
const (
	MessageTypeCallConnectRequest MessageType = 1
	MessageTypeCallConnectAck     MessageType = 2
	MessageTypeCallConnectNak     MessageType = 3
	MessageTypeCallConnected      MessageType = 4
	MessageTypeCallAbort          MessageType = 5
	MessageTypeCallDisconnect     MessageType = 6
	MessageTypeCallDisconnectAck  MessageType = 7
	MessageTypeEchoRequest        MessageType = 8
	MessageTypeEchoResponse       MessageType = 9
)

func (k MessageType) String() string {
	switch k {
	case MessageTypeCallConnectRequest:
		return "CallConnectRequest"
	case MessageTypeCallConnectAck:
		return "CallConnectAck"
	case MessageTypeCallConnectNak:
		return "CallConnectNak"
	case MessageTypeCallConnected:
		return "CallConnected"
	case MessageTypeCallAbort:
		return "CallAbort"
	case MessageTypeCallDisconnect:
		return "CallDisconnect"
	case MessageTypeCallDisconnectAck:
		return "CallDisconnectAck"
	case MessageTypeEchoRequest:
		return "EchoRequest"
	case MessageTypeEchoResponse:
		return "EchoResponse"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

type sstpControlHeader struct {
	sstpHeader
	MessageType      MessageType
	AttributesLength uint16
	Attributes       []sstpAttribute
}

// AttributeID is the type of attribute this attribute is
type AttributeID uint8

// Constants for AttributeID values
const (
	AttributeIDEncapsulatedProtocolID AttributeID = 1
	AttributeIDStatusInfo             AttributeID = 2
	AttributeIDCryptoBinding          AttributeID = 3
	AttributeIDCryptoBindingReq       AttributeID = 4
)

func (k AttributeID) String() string {
	switch k {
	case AttributeIDEncapsulatedProtocolID:
		return "EncapsulatedProtocolID"
	case AttributeIDStatusInfo:
		return "StatusInfo"
	case AttributeIDCryptoBinding:
		return "CryptoBinding"
	case AttributeIDCryptoBindingReq:
		return "CryptoBindingReq"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

type sstpAttribute struct {
	Reserved    byte
	AttributeID AttributeID
	Length      uint16
	Data        []byte
}

type sstpDataHeader struct {
	sstpHeader
	Data []byte
}

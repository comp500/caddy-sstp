package ppp

import (
	"encoding/binary"
	"fmt"
	"log"
)

type lcpProtocol struct{}

// TODO: move to cp.go?
// controlCode is the LCP/IPCP/CCP control protocol message code of this packet
type controlCode uint16

// Constants for controlCode values
// See https://www.iana.org/assignments/ppp-numbers/ppp-numbers.xml for the full list
// Don't use iota, as each number MUST be equal to given
const (
	controlCodeConfigureRequest controlCode = 1
	controlCodeConfigureAck     controlCode = 2
	controlCodeConfigureNak     controlCode = 3
	controlCodeConfigureReject  controlCode = 4
	controlCodeTerminateRequest controlCode = 5
	controlCodeTerminateAck     controlCode = 6
	controlCodeReject           controlCode = 7
	controlCodeProtocolReject   controlCode = 8
	controlCodeEchoRequest      controlCode = 9
	controlCodeEchoReply        controlCode = 10
	controlCodeDiscardRequest   controlCode = 11
	// implement Identification and Time-Remaining?
	// implement CCP?
)

func (k controlCode) String() string {
	switch k {
	case controlCodeConfigureRequest:
		return "Configure-Request"
	case controlCodeConfigureAck:
		return "Configure-Ack"
	case controlCodeConfigureNak:
		return "Configure-Nak"
	case controlCodeConfigureReject:
		return "Configure-Reject"
	case controlCodeTerminateRequest:
		return "Terminate-Request"
	case controlCodeTerminateAck:
		return "Terminate-Ack"
	case controlCodeReject:
		return "Code-Reject"
	case controlCodeProtocolReject:
		return "Protocol-Reject"
	case controlCodeEchoRequest:
		return "Echo-Request"
	case controlCodeEchoReply:
		return "Echo-Reply"
	case controlCodeDiscardRequest:
		return "Discard-Request"
	default:
		return fmt.Sprintf("Unknown (%d)", k)
	}
}

// lcpOption is a LCP Configuration Option
type lcpOption uint16

// Constants for lcpOption values
const (
	lcpOptionMRU             lcpOption = 1 // Maximum-Recieve-Unit
	lcpOptionAuthProtocol    lcpOption = 3
	lcpOptionQualityProtocol lcpOption = 4
	lcpOptionMagicNumber     lcpOption = 5
	lcpOptionPFC             lcpOption = 7 // Protocol-Field-Compression
	lcpOptionACFC            lcpOption = 8 // Address-and-Control-Field-Compression
)

func (k lcpOption) String() string {
	switch k {
	case lcpOptionMRU:
		return "Maximum-Receive-Unit"
	case lcpOptionAuthProtocol:
		return "Authentication-Protocol"
	case lcpOptionQualityProtocol:
		return "Quality-Protocol"
	case lcpOptionMagicNumber:
		return "Magic-Number"
	case lcpOptionPFC:
		return "Protocol-Field-Compression"
	case lcpOptionACFC:
		return "Address-and-Control-Field-Compression"
	default:
		return fmt.Sprintf("Unknown (%d)", k)
	}
}

type lcpOptionData struct {
	option lcpOption
	data   []byte
}

type lcpPacket struct {
	code       controlCode
	identifier int
}

type lcpConfigurePacket struct {
	lcpPacket
	options []lcpOptionData
}

type lcpTerminatePacket struct {
	lcpPacket
	data []byte
}

type lcpCodeRejectPacket struct {
	lcpPacket
	rejectedData []byte
}

type lcpProtocolRejectPacket struct {
	lcpPacket
	rejectedProtocol protocolType
	rejectedData     []byte
}

type lcpEchoPacket struct {
	lcpPacket
	magicNumber int32
	data        []byte
}

type lcpDiscardPacket struct {
	lcpPacket
	magicNumber int32
	data        []byte
}

// Write data from higher layers into LCP
func (p *lcpProtocol) writeData(data []byte, h *controlProtocolHelper) (int, error) {
	controlCode := controlCode(data[0])
	// TODO implement identifier?
	packetLength := binary.BigEndian.Uint16(data[2:4])

	// Shift data along, trim to packetLength
	data = data[4:packetLength]

	switch controlCode {
	case controlCodeConfigureRequest:
		log.Printf("It's a Configure-Request")
	default:
		log.Printf("%s not implemented", controlCode)
	}
	// Must silently discard any Discard-Request packets

	return len(data), nil
}

func (p *lcpProtocol) writePacket(packet lcpPacket, data []byte) error {
	return nil
}

func (p *lcpProtocol) writeConfigurePacket(packet lcpConfigurePacket, data []byte) error {
	return nil
}

func (p *lcpProtocol) writeTerminatePacket(packet lcpTerminatePacket, data []byte) error {
	return nil
}

func (p *lcpProtocol) writeCodeRejectPacket(packet lcpCodeRejectPacket, data []byte) error {
	return nil
}

func (p *lcpProtocol) writeProtocolRejectPacket(packet lcpProtocolRejectPacket, data []byte) error {
	return nil
}

func (p *lcpProtocol) writeEchoPacket(packet lcpEchoPacket, data []byte) error {
	return nil
}

func (p *lcpProtocol) sendConfigureRequest(h *controlProtocolHelper) error {
	h.configureCount--
	return nil
}

func (p *lcpProtocol) sendConfigureAck(h *controlProtocolHelper) error {
	return nil
}

func (p *lcpProtocol) sendConfigureNak(h *controlProtocolHelper) error {
	return nil
}

func (p *lcpProtocol) sendTerminateRequest(h *controlProtocolHelper) error {
	h.terminateCount--
	return nil
}

func (p *lcpProtocol) sendTerminateAck(h *controlProtocolHelper) error {
	return nil
}

func (p *lcpProtocol) sendCodeReject(h *controlProtocolHelper) error {
	return nil
}

func (p *lcpProtocol) sendEchoReply(h *controlProtocolHelper) error {
	return nil
}

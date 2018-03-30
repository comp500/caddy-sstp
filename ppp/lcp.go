package ppp

import (
	"encoding/binary"
	"fmt"
	"log"
)

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
	lcpOptionMRU          lcpOption = 1 // Maximum-Recieve-Unit
	lcpOptionAuthProtocol lcpOption = 3
	lcpOptionMagicNumber  lcpOption = 4
	lcpOptionPFC          lcpOption = 7 // Protocol-Field-Compression
	lcpOptionACFC         lcpOption = 8 // Address-and-Control-Field-Compression
)

func (k lcpOption) String() string {
	switch k {
	case lcpOptionMRU:
		return "Maximum-Receive-Unit"
	case lcpOptionAuthProtocol:
		return "Authentication-Protocol"
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

func (p *nativeConnection) parseLCP(data []byte) error {
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

	return nil
}

type lcpProtocol struct{}

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

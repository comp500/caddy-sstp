package ppp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

// This file manages pppd connections for the native (pure Go) connection type.
type nativeConnection struct {
	Config
	linkStatus     linkStatus
	firstFrameSent bool
	hasBeenClosed  bool
	// Indicates whether Address-and-Control-Field-Compression is applied
	acfcApplied bool
	// Indicates whether Protocol-Field-Compression is applied
	pfcApplied bool
}

func (p *nativeConnection) Write(data []byte) (int, error) {
	if p.hasBeenClosed {
		return 0, errors.New("ppp write after close")
	}
	err := parsePPP(data, p)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

func (p *nativeConnection) Close() error {
	p.linkStatus = linkStatusDead
	p.hasBeenClosed = true
	return nil
}

func (p *nativeConnection) start() error {
	echoRequest := [...]byte{0xff, 0x03, 0xc0, 0x21, 0x09, 0x00, 0x00, 0x08, 0x58, 0xa5, 0xe7, 0xc2}
	p.DestWriter.Write(echoRequest[:])
	return nil
}

// linkStatus is the current status of the PPP connection
type linkStatus int

// Constants for linkStatus values
const (
	linkStatusDead linkStatus = iota
	linkStatusEstablish
	linkStatusAuthenticate
	linkStatusNetwork
	linkStatusTerminate
)

// protocolType is the protocol that this PPP packet uses
type protocolType uint16

// Constants for protocolType values
const (
	protocolTypeLCP  protocolType = 0xC021
	protocolTypePAP  protocolType = 0xC023
	protocolTypeCHAP protocolType = 0xC223
	protocolTypeIPCP protocolType = 0x8021
	protocolTypeIP   protocolType = 0x0021
	protocolTypeCCP  protocolType = 0x80fd
)

// accessControlFields are the Address and Control fields, often interpreted as a unknown protocol
// ACFC (Address-and-Control-Field-Compression) removes this
const accessControlFields = 0xff03

func (k protocolType) String() string {
	switch k {
	case protocolTypeLCP:
		return "LCP"
	case protocolTypePAP:
		return "PAP"
	case protocolTypeCHAP:
		return "CHAP"
	case protocolTypeIPCP:
		return "IPCP"
	case protocolTypeIP:
		return "IP"
	case protocolTypeCCP:
		return "CCP"
	case accessControlFields:
		return "Unknown (Access/Control fields)"
	default:
		return fmt.Sprintf("Unknown (%d)", k)
	}
}

func parsePPP(data []byte, p *nativeConnection) error {
	// TODO: parse packets for *every* protocol
	if !p.acfcApplied {
		// If address and control field is not compressed, remove it
		data = data[2:]
	}

	var protocolNumber protocolType

	// If protocol field is compressed, uncompress protocolNumber
	// Test if LSB is set
	if p.pfcApplied && (data[0]&1 == 1) {
		protocolNumber = protocolType(data[0])
		// Shift data along
		data = data[1:]
	} else {
		protocolNumber = protocolType(binary.BigEndian.Uint16(data[0:2]))
		// Shift data along
		data = data[2:]
	}

	if p.linkStatus == linkStatusDead {
		// Data is received, wake up
		p.linkStatus = linkStatusEstablish
	}

	if p.linkStatus == linkStatusEstablish {
		if protocolNumber == protocolTypeLCP {
			log.Print("LCP")
			return parseLCP(data, p)
		}
		log.Print("Discarding packet")
		// silently discard, only allow LCP
	}

	if p.linkStatus == linkStatusAuthenticate {
		switch protocolNumber {
		case protocolTypeLCP:
			log.Print("LCP")
			return parseLCP(data, p)
		case protocolTypePAP:
			log.Print("PAP")
		case protocolTypeCHAP:
			log.Print("CHAP")
		default:
			log.Print("Discarding packet")
			// silently discard
		}
	}

	if p.linkStatus == linkStatusNetwork {
		switch protocolNumber {
		case protocolTypeIP:
			log.Print("IP")
		// for protocols known, but not used in this phase
		case protocolTypePAP, protocolTypeCHAP:
			log.Print("Discarding packet")
			// silently discard
		case protocolTypeLCP:
			log.Print("LCP")
			return parseLCP(data, p)
		case protocolTypeIPCP:
			log.Print("IPCP")
		case protocolTypeCCP:
			log.Print("CCP")
		default:
			// TODO send LCP Protocol Reject
			log.Print("Unknown protocol, rejecting")
		}
	}

	if p.linkStatus == linkStatusTerminate {
		if protocolNumber == protocolTypeLCP {
			log.Print("LCP")
			return parseLCP(data, p)
		}
		log.Print("Discarding packet")
		// silently discard, only allow LCP
	}

	return nil
}

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

func parseLCP(data []byte, p *nativeConnection) error {
	log.Printf("%s", controlCode(data[0]))
	return nil
}

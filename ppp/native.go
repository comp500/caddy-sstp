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
	Vnat           bool
	firstFrameSent bool
	hasBeenClosed  bool
	acfcApplied    bool // Indicates whether Address-and-Control-Field-Compression is applied
	pfcApplied     bool // Indicates whether Protocol-Field-Compression is applied
	lcpState       lcpState
}

func (p *nativeConnection) Write(data []byte) (int, error) {
	if p.hasBeenClosed {
		return 0, errors.New("ppp write after close")
	}
	err := p.parsePPP(data)
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

func (p *nativeConnection) parsePPP(data []byte) error {
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
			return p.parseLCP(data)
		}
		log.Print("Discarding packet")
		// silently discard, only allow LCP
	}

	if p.linkStatus == linkStatusAuthenticate {
		switch protocolNumber {
		case protocolTypeLCP:
			log.Print("LCP")
			return p.parseLCP(data)
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
			return p.parseLCP(data)
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
			return p.parseLCP(data)
		}
		log.Print("Discarding packet")
		// silently discard, only allow LCP
	}

	return nil
}

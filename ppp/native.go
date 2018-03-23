package ppp

import (
	"encoding/binary"
	"fmt"
	"log"
)

// This file manages pppd connections for the native (pure Go) connection type.
type nativeConnection struct {
	Config
	linkStatus     linkStatus
	firstFrameSent bool
}

func (p *nativeConnection) Write(data []byte) (int, error) {
	parsePPP(data, p)
	return len(data), nil
}

func (p *nativeConnection) Close() error {
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
	linkStatusDead         = 0
	linkStatusEstablish    = 1
	linkStatusAuthenticate = 2
	linkStatusNetwork      = 3
	linkStatusTerminate    = 4
)

// protocolType is the protocol that this packet uses
type protocolType uint16

// Constants for protocolType values
const (
	protocolTypeLCP  = 0xC021
	protocolTypePAP  = 0xC023
	protocolTypeCHAP = 0xC223
	protocolTypeIPCP = 0x8021
	protocolTypeIP   = 0x0021
	protocolTypeCCP  = 0x80fd
)

// HDLC-like header flag, given at the start of some packets for each PPP connection, both directions
// TODO: When does this happen, when should it be applied?
const hdlcFlag = 0xff03

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
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

func parsePPP(data []byte, p *nativeConnection) {
	// TODO: parse packets for *every* protocol
	protocolNumber := binary.BigEndian.Uint16(data[0:2])

	if protocolNumber == hdlcFlag {
		// If HDLC flag found, remove it and parse again
		data = data[2:]
		protocolNumber = binary.BigEndian.Uint16(data[0:2])
	}

	if p.linkStatus == linkStatusDead {
		// Data is received, wake up
		p.linkStatus = linkStatusEstablish
	}

	if p.linkStatus == linkStatusEstablish {
		if protocolNumber == protocolTypeLCP {
			log.Print("LCP")
			return
		}
		log.Print("Discarding packet")
		// silently discard, only allow LCP
	}

	if p.linkStatus == linkStatusAuthenticate {
		switch protocolNumber {
		case protocolTypeLCP:
			log.Print("LCP")
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
		if data[0] == protocolTypeIP {
			log.Print("IP")
			return
		}

		switch protocolNumber {
		// for protocols known, but not used in this phase
		case protocolTypePAP, protocolTypeCHAP:
			log.Print("Discarding packet")
			// silently discard
		case protocolTypeLCP:
			log.Print("LCP")
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
			return
		}
		log.Print("Discarding packet")
		// silently discard, only allow LCP
	}
}

func sendLCPEchoRequest() {

}

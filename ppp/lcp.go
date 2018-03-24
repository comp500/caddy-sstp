package ppp

import (
	"encoding/binary"
	"errors"
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

// lcpState is the current status of the LCP negotiation automaton
type lcpState int

// Constants for lcpState values
const (
	lcpStateInitial lcpState = iota
	lcpStateStarting
	lcpStateClosed
	lcpStateStopped
	lcpStateClosing
	lcpStateStopping
	lcpStateReqSent
	lcpStateAckReceived
	lcpStateAckSent
	lcpStateOpened
)

// ErrLcpAutomaton is an internal error in the LCP automaton
var ErrLcpAutomaton = errors.New("Invalid LCP automaton state")

// TODO: make this configurable
const (
	lcpMaxTerminate = 2
	lcpMaxConfigure = 10
	lcpMaxFailure   = 5
)

// LCP automaton events, see RFC1661 section 4.1

func (p *nativeConnection) lcpUp() error {
	switch p.lcpState {
	case lcpStateInitial:
		p.lcpState = lcpStateClosed
	case lcpStateStarting:
		p.lcpRestartCount = lcpMaxConfigure
		err := p.lcpSendConfigureRequest()
		if err != nil {
			return err
		}
		p.lcpState = lcpStateReqSent
	default:
		return ErrLcpAutomaton
	}
	return nil
}

func (p *nativeConnection) lcpDown() error {
	switch p.lcpState {
	case lcpStateClosed:
		p.lcpState = lcpStateInitial
	case lcpStateStopped:
		// TODO: THIS-LAYER-STARTED
		// - should signal to lower layers to start
		// - once started, lcpUp should be called
		p.lcpState = lcpStateStarting
	case lcpStateClosing:
		p.lcpState = lcpStateInitial
	case lcpStateStopping, lcpStateReqSent, lcpStateAckReceived, lcpStateAckSent:
		p.lcpState = lcpStateStarting
	case lcpStateOpened:
		// TODO: THIS-LAYER-DOWN
		// - should signal to upper layers that it is leaving Opened state
		// - e.g. signal Down event to NCP/Auth/LQP
		p.lcpState = lcpStateStarting
	default:
		return ErrLcpAutomaton
	}
	return nil
}

func (p *nativeConnection) lcpOpen() error {
	switch p.lcpState {
	case lcpStateInitial:
		// TODO: THIS-LAYER-STARTED
		// - should signal to lower layers to start
		// - once started, lcpUp should be called
		p.lcpState = lcpStateStarting
	case lcpStateStarting:
		// Do nothing, 1 -> 1
	case lcpStateClosed:
		p.lcpRestartCount = lcpMaxConfigure
		err := p.lcpSendConfigureRequest()
		if err != nil {
			return err
		}
		p.lcpState = lcpStateReqSent
	case lcpStateStopped:
		// Do nothing, 3 -> 3
		// restart?
	case lcpStateClosing:
		p.lcpState = lcpStateStopping
		// restart?
	case lcpStateStopping:
		// Do nothing, 5 -> 5
		// restart?
	case lcpStateReqSent, lcpStateAckReceived, lcpStateAckSent:
		// Do nothing
	case lcpStateOpened:
		// Do nothing, 9 -> 9
		// restart?
	default:
		return ErrLcpAutomaton
	}
	return nil
}

func (p *nativeConnection) lcpClose() error {
	switch p.lcpState {
	case lcpStateInitial:
		// Do nothing, 0 -> 0
	case lcpStateStarting:
		// TODO: THIS-LAYER-FINISHED
		// - advance to Link Dead phase in PPP
		p.lcpState = lcpStateInitial
	case lcpStateClosed:
		// Do nothing, 2 -> 2
	case lcpStateStopped:
		p.lcpState = lcpStateClosed
	case lcpStateClosing:
		// Do nothing, 4 -> 4
	case lcpStateStopping:
		p.lcpState = lcpStateClosing
	case lcpStateOpened:
		// TODO: THIS-LAYER-DOWN
		// - should signal to upper layers that it is leaving Opened state
		// - e.g. signal Down event to NCP/Auth/LQP
		fallthrough
	case lcpStateReqSent, lcpStateAckReceived, lcpStateAckSent:
		p.lcpRestartCount = lcpMaxFailure
		err := p.lcpSendTerminateRequest()
		if err != nil {
			return err
		}
		p.lcpState = lcpStateClosing
	default:
		return ErrLcpAutomaton
	}
	return nil
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

func (p *nativeConnection) lcpSendConfigureRequest() error {
	p.lcpRestartCount--
	return nil
}

func (p *nativeConnection) lcpSendTerminateRequest() error {
	p.lcpRestartCount--
	return nil
}

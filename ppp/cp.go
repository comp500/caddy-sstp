package ppp

import (
	"errors"
)

// Generic Control Protocol interface with helper methods for automatons
type controlProtocol interface {
	sendConfigureRequest(*controlProtocolHelper) error
	sendTerminateRequest(*controlProtocolHelper) error
}

type controlProtocolHelper struct {
	controlProtocol
	state        cpState
	restartCount int
	failureCount int
}

// cpState is the current status of the CP negotiation automaton
type cpState int

// Constants for cpState values
const (
	cpStateInitial cpState = iota
	cpStateStarting
	cpStateClosed
	cpStateStopped
	cpStateClosing
	cpStateStopping
	cpStateReqSent
	cpStateAckReceived
	cpStateAckSent
	cpStateOpened
)

// ErrCpAutomaton is an internal error in the Control Protocol automaton
var ErrCpAutomaton = errors.New("Invalid Control Protocol automaton state")

// TODO: make this configurable
const (
	cpMaxTerminate = 2
	cpMaxConfigure = 10
	cpMaxFailure   = 5
)

// CP automaton events, see RFC1661 section 4.1

func (p *controlProtocolHelper) lcpUp() error {
	switch p.state {
	case cpStateInitial:
		p.state = cpStateClosed
	case cpStateStarting:
		p.restartCount = cpMaxConfigure
		err := p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) lcpDown() error {
	switch p.state {
	case cpStateClosed:
		p.state = cpStateInitial
	case cpStateStopped:
		// TODO: THIS-LAYER-STARTED
		// - should signal to lower layers to start
		// - once started, lcpUp should be called
		p.state = cpStateStarting
	case cpStateClosing:
		p.state = cpStateInitial
	case cpStateStopping, cpStateReqSent, cpStateAckReceived, cpStateAckSent:
		p.state = cpStateStarting
	case cpStateOpened:
		// TODO: THIS-LAYER-DOWN
		// - should signal to upper layers that it is leaving Opened state
		// - e.g. signal Down event to NCP/Auth/LQP
		p.state = cpStateStarting
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) lcpOpen() error {
	switch p.state {
	case cpStateInitial:
		// TODO: THIS-LAYER-STARTED
		// - should signal to lower layers to start
		// - once started, lcpUp should be called
		p.state = cpStateStarting
	case cpStateStarting:
		// Do nothing, 1 -> 1
	case cpStateClosed:
		p.restartCount = cpMaxConfigure
		err := p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
	case cpStateStopped:
		// Do nothing, 3 -> 3
		// restart?
	case cpStateClosing:
		p.state = cpStateStopping
		// restart?
	case cpStateStopping:
		// Do nothing, 5 -> 5
		// restart?
	case cpStateReqSent, cpStateAckReceived, cpStateAckSent:
		// Do nothing
	case cpStateOpened:
		// Do nothing, 9 -> 9
		// restart?
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) lcpClose() error {
	switch p.state {
	case cpStateInitial:
		// Do nothing, 0 -> 0
	case cpStateStarting:
		// TODO: THIS-LAYER-FINISHED
		// - advance to Link Dead phase in PPP
		p.state = cpStateInitial
	case cpStateClosed:
		// Do nothing, 2 -> 2
	case cpStateStopped:
		p.state = cpStateClosed
	case cpStateClosing:
		// Do nothing, 4 -> 4
	case cpStateStopping:
		p.state = cpStateClosing
	case cpStateOpened:
		// TODO: THIS-LAYER-DOWN
		// - should signal to upper layers that it is leaving Opened state
		// - e.g. signal Down event to NCP/Auth/LQP
		fallthrough
	case cpStateReqSent, cpStateAckReceived, cpStateAckSent:
		p.restartCount = cpMaxFailure
		err := p.sendTerminateRequest(p)
		if err != nil {
			return err
		}
		p.state = cpStateClosing
	default:
		return ErrCpAutomaton
	}
	return nil
}

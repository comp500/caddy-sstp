package ppp

import (
	"errors"
	"time"
)

// Generic Control Protocol interface with helper methods for automatons
type controlProtocol interface {
	writeData([]byte, *controlProtocolHelper) (int, error)
	sendConfigureRequest(*controlProtocolHelper) error
	sendConfigureAck(*controlProtocolHelper) error
	sendConfigureNak(*controlProtocolHelper) error
	sendTerminateRequest(*controlProtocolHelper) error
	sendTerminateAck(*controlProtocolHelper) error
	sendCodeReject(*controlProtocolHelper) error
	sendEchoReply(*controlProtocolHelper) error
}

type controlProtocolHelper struct {
	controlProtocol
	state               cpState
	configureCount      int
	terminateCount      int
	restartTimer        *time.Timer // TODO read from the timer
	restartTimerExpired bool
	failureCount        int
}

// Implement io.Writer
func (p *controlProtocolHelper) Write(data []byte) (int, error) {
	return p.writeData(data, p)
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

// TODO: make these configurable
const (
	cpMaxTerminate = 2
	cpMaxConfigure = 10
	cpMaxFailure   = 5
	cpTimerLength  = 3 * time.Second
)

// CP automaton actions

// This-Layer-Up
func (p *controlProtocolHelper) tlu() error {
	return nil
}

// This-Layer-Down
func (p *controlProtocolHelper) tld() error {
	// TODO: THIS-LAYER-DOWN
	// - should signal to upper layers that it is leaving Opened state
	// - e.g. signal Down event to NCP/Auth/LQP
	return nil
}

// This-Layer-Started
func (p *controlProtocolHelper) tls() error {
	// TODO: THIS-LAYER-STARTED
	// - should signal to lower layers to start
	// - once started, lcpUp should be called
	return nil
}

// This-Layer-Finished
func (p *controlProtocolHelper) tlf() error {
	// TODO: THIS-LAYER-FINISHED
	// - advance to Link Dead phase in PPP
	return nil
}

// Initialize-Restart-Count
func (p *controlProtocolHelper) irc(isTerminate bool) error {
	if isTerminate {
		p.terminateCount = cpMaxTerminate
	} else {
		p.configureCount = cpMaxConfigure
	}
	p.resetTimer()
	return nil
}

// Zero-Restart-Count
func (p *controlProtocolHelper) zrc() error {
	return nil
}

// CP automaton events, see RFC1661 section 4.1

func (p *controlProtocolHelper) Up() error {
	switch p.state {
	case cpStateInitial:
		p.state = cpStateClosed
	case cpStateStarting:
		err := p.irc(false)
		if err != nil {
			return err
		}
		// TODO: store corresponding request for timer?
		err = p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) Down() error {
	switch p.state {
	case cpStateClosed:
		p.state = cpStateInitial
	case cpStateStopped:
		err := p.tls()
		if err != nil {
			return err
		}
		p.state = cpStateStarting
	case cpStateClosing:
		p.state = cpStateInitial
	case cpStateStopping, cpStateReqSent, cpStateAckReceived, cpStateAckSent:
		p.state = cpStateStarting
	case cpStateOpened:
		err := p.tld()
		if err != nil {
			return err
		}
		p.state = cpStateStarting
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) Open() error {
	switch p.state {
	case cpStateInitial:
		err := p.tls()
		if err != nil {
			return err
		}
		p.state = cpStateStarting
	case cpStateStarting:
		// Do nothing, 1 -> 1
	case cpStateClosed:
		err := p.irc(false)
		if err != nil {
			return err
		}
		// TODO: store corresponding request for timer?
		err = p.sendConfigureRequest(p)
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

func (p *controlProtocolHelper) Close() error {
	switch p.state {
	case cpStateInitial:
		// Do nothing, 0 -> 0
	case cpStateStarting:
		err := p.tlf()
		if err != nil {
			return err
		}
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
		err := p.tld()
		if err != nil {
			return err
		}
		fallthrough
	case cpStateReqSent, cpStateAckReceived, cpStateAckSent:
		err := p.irc(true)
		if err != nil {
			return err
		}
		// TODO: store corresponding request for timer?
		err = p.sendTerminateRequest(p)
		if err != nil {
			return err
		}
		p.state = cpStateClosing
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) resetTimer() {
	if p.restartTimer != nil {
		p.restartTimer = time.NewTimer(cpTimerLength)
	} else {
		if !p.restartTimerExpired && !p.restartTimer.Stop() {
			<-p.restartTimer.C
		}
		p.restartTimer.Reset(cpTimerLength)
	}
}

func (p *controlProtocolHelper) timeoutTriggered() error {
	// TODO: which counter do we check?
	// should we have multiple timers for configure/terminate?

	if p.configureCount > 0 {
		switch p.state {
		case cpStateClosing, cpStateStopping:
			err := p.sendTerminateRequest(p)
			if err != nil {
				return err
			}
		case cpStateAckReceived:
			p.state = cpStateReqSent
			fallthrough
		case cpStateReqSent, cpStateAckSent:
			err := p.sendConfigureRequest(p)
			if err != nil {
				return err
			}
		default:
			return ErrCpAutomaton
		}
	} else {
		switch p.state {
		case cpStateClosing:
			err := p.tlf()
			if err != nil {
				return err
			}
			p.state = cpStateClosed
		case cpStateReqSent, cpStateAckReceived, cpStateAckSent:
			// passive?
			fallthrough
		case cpStateStopping:
			err := p.tlf()
			if err != nil {
				return err
			}
			p.state = cpStateStopped
		default:
			return ErrCpAutomaton
		}
	}
	return nil
}

func (p *controlProtocolHelper) receiveGoodConfigureRequest() error {
	switch p.state {
	case cpStateClosed:
		err := p.sendTerminateAck(p)
		if err != nil {
			return err
		}
	case cpStateStopped:
		err := p.irc(false)
		if err != nil {
			return err
		}
		err = p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		err = p.sendConfigureAck(p)
		if err != nil {
			return err
		}
		p.state = cpStateAckSent
	case cpStateClosing, cpStateStopping:
		// Do nothing
	case cpStateReqSent:
		err := p.sendConfigureAck(p)
		if err != nil {
			return err
		}
		p.state = cpStateAckSent
	case cpStateAckReceived:
		err := p.sendConfigureAck(p)
		if err != nil {
			return err
		}
		err = p.tlu()
		if err != nil {
			return err
		}
		p.state = cpStateOpened
	case cpStateAckSent:
		err := p.sendConfigureAck(p)
		if err != nil {
			return err
		}
	case cpStateOpened:
		err := p.tld()
		if err != nil {
			return err
		}
		err = p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		err = p.sendConfigureAck(p)
		if err != nil {
			return err
		}
		p.state = cpStateAckSent
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) receiveBadConfigureRequest() error {
	switch p.state {
	case cpStateClosed:
		err := p.sendTerminateAck(p)
		if err != nil {
			return err
		}
	case cpStateStopped:
		err := p.irc(false)
		if err != nil {
			return err
		}
		err = p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		err = p.sendConfigureNak(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
	case cpStateClosing, cpStateStopping:
		// Do nothing
	case cpStateReqSent:
		err := p.sendConfigureNak(p)
		if err != nil {
			return err
		}
	case cpStateAckReceived:
		err := p.sendConfigureNak(p)
		if err != nil {
			return err
		}
	case cpStateAckSent:
		err := p.sendConfigureNak(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
	case cpStateOpened:
		err := p.tld()
		if err != nil {
			return err
		}
		err = p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		err = p.sendConfigureNak(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) receiveConfigureAck() error {
	switch p.state {
	case cpStateClosed:
		err := p.sendTerminateAck(p)
		if err != nil {
			return err
		}
	case cpStateStopped:
		err := p.sendTerminateAck(p)
		if err != nil {
			return err
		}
	case cpStateClosing, cpStateStopping:
		// Do nothing
	case cpStateReqSent:
		err := p.irc(false)
		if err != nil {
			return err
		}
		p.state = cpStateAckReceived
	case cpStateAckReceived:
		err := p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
		// crossed?
	case cpStateAckSent:
		err := p.irc(false)
		if err != nil {
			return err
		}
		err = p.tlu()
		if err != nil {
			return err
		}
		p.state = cpStateOpened
	case cpStateOpened:
		err := p.tld()
		if err != nil {
			return err
		}
		err = p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
		// crossed?
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) receiveConfigureNak() error {
	switch p.state {
	case cpStateClosed:
		err := p.sendTerminateAck(p)
		if err != nil {
			return err
		}
	case cpStateStopped:
		err := p.sendTerminateAck(p)
		if err != nil {
			return err
		}
	case cpStateClosing, cpStateStopping:
		// Do nothing
	case cpStateReqSent:
		err := p.irc(false)
		if err != nil {
			return err
		}
		err = p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
	case cpStateAckReceived:
		err := p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
		// crossed?
	case cpStateAckSent:
		err := p.irc(false)
		if err != nil {
			return err
		}
		err = p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
	case cpStateOpened:
		err := p.tld()
		if err != nil {
			return err
		}
		err = p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
		// crossed?
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) receiveTerminateRequest() error {
	switch p.state {
	case cpStateClosed, cpStateStopped, cpStateClosing, cpStateStopping, cpStateReqSent:
		err := p.sendTerminateAck(p)
		if err != nil {
			return err
		}
	case cpStateAckReceived, cpStateAckSent:
		err := p.sendTerminateAck(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
	case cpStateOpened:
		err := p.tld()
		if err != nil {
			return err
		}
		err = p.zrc()
		if err != nil {
			return err
		}
		err = p.sendTerminateAck(p)
		if err != nil {
			return err
		}
		p.state = cpStateStopping
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) receiveTerminateAck() error {
	switch p.state {
	case cpStateClosed, cpStateStopped:
		// Do nothing
	case cpStateClosing:
		err := p.tlf()
		if err != nil {
			return err
		}
		p.state = cpStateClosed
	case cpStateStopping:
		err := p.tlf()
		if err != nil {
			return err
		}
		p.state = cpStateStopped
	case cpStateReqSent, cpStateAckSent:
		// Do nothing
	case cpStateAckReceived:
		p.state = cpStateReqSent
	case cpStateOpened:
		err := p.tld()
		if err != nil {
			return err
		}
		err = p.sendConfigureRequest(p)
		if err != nil {
			return err
		}
		p.state = cpStateReqSent
	default:
		return ErrCpAutomaton
	}
	return nil
}

func (p *controlProtocolHelper) receiveUnknownCode() error {
	switch p.state {
	case cpStateClosed, cpStateStopped, cpStateClosing, cpStateStopping, cpStateReqSent, cpStateAckReceived, cpStateAckSent, cpStateOpened:
		err := p.sendCodeReject(p)
		if err != nil {
			return err
		}
	default:
		return ErrCpAutomaton
	}
	return nil
}

// or Receive-Protocol-Reject
func (p *controlProtocolHelper) receiveCodeRejectPermitted() error {
	switch p.state {
	case cpStateClosed, cpStateStopped, cpStateClosing, cpStateStopping, cpStateReqSent, cpStateAckSent, cpStateOpened:
		// Do nothing
	case cpStateAckReceived:
		p.state = cpStateReqSent
	default:
		return ErrCpAutomaton
	}
	return nil
}

// or Receive-Protocol-Reject
func (p *controlProtocolHelper) receiveCodeRejectCatastrophic() error {
	switch p.state {
	case cpStateClosed, cpStateStopped:
		err := p.tlf()
		if err != nil {
			return err
		}
	case cpStateClosing:
		err := p.tlf()
		if err != nil {
			return err
		}
		p.state = cpStateClosed
	case cpStateStopping, cpStateReqSent, cpStateAckReceived, cpStateAckSent:
		err := p.tlf()
		if err != nil {
			return err
		}
		p.state = cpStateStopped
	case cpStateOpened:
		err := p.tld()
		if err != nil {
			return err
		}
		err = p.irc(true)
		if err != nil {
			return err
		}
		err = p.sendTerminateRequest(p)
		if err != nil {
			return err
		}
		p.state = cpStateStopping
	default:
		return ErrCpAutomaton
	}
	return nil
}

// or Receive-Echo-Reply
// or Receive-Discard-Request
func (p *controlProtocolHelper) receiveEchoRequest() error {
	switch p.state {
	case cpStateClosed, cpStateStopped, cpStateClosing, cpStateStopping, cpStateReqSent, cpStateAckReceived, cpStateAckSent:
		// Do nothing
	case cpStateOpened:
		p.sendEchoReply(p)
	default:
		return ErrCpAutomaton
	}
	return nil
}

package ppp

import (
	"io"
	"log"
	"net"
	"os/exec"
)

// TODO: generalise, to custom implementation of ppp

// PppdInstance represents an instance of pppd
type PppdInstance struct {
	commandInst *exec.Cmd
	stdin       io.WriteCloser
	unescaper   pppUnescaper
	args        []string
	IsStarted   bool
}

// StartPppd starts pppd
func StartPppd(pppdInstance *PppdInstance, conn net.Conn) error {
	args := append([]string{"notty", "file", "/etc/ppp/options.sstpd"}, pppdInstance.args...)
	pppdCmd := exec.Command("pppd", args...)
	pppdIn, err := pppdCmd.StdinPipe()
	if err != nil {
		return err
	}
	pppdCmd.Stdout = pppdInstance.unescaper
	err = pppdCmd.Start()
	if err != nil {
		return err
	}
	pppdInstance.commandInst = pppdCmd
	pppdInstance.stdin = pppdIn
	// TODO: pipe stderr to log?

	pppdInstance.IsStarted = true

	// TODO: remove this?
	go func() {
		defer log.Print("pppd disconnected")
		pppdCmd.Wait()
	}()

	return nil
}

// NewPppdInstance creates a new pppd instance object
func NewPppdInstance(outputWriter io.Writer, args []string) PppdInstance {
	return PppdInstance{nil, nil, newUnescaper(outputWriter), args, false}
}

// Kill kills pppd
func (p PppdInstance) Kill() error {
	if p.commandInst != nil {
		err := p.commandInst.Process.Kill()
		if err != nil {
			return err
		}
		p.commandInst = nil
	}
	p.IsStarted = false
	return nil
}

func (p PppdInstance) Write(data []byte) (int, error) {
	return p.stdin.Write(pppEscape(data))
}

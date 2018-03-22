package ppp

import (
	"io"
	"log"
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

// NewPppdInstance creates a new pppd instance object
func NewPppdInstance(outputWriter io.Writer, args []string) PppdInstance {
	return PppdInstance{nil, nil, newUnescaper(outputWriter), args, false}
}

// Start starts pppd
func (p PppdInstance) Start() error {
	args := append([]string{"notty", "file", "/etc/ppp/options.sstpd"}, p.args...)
	pppdCmd := exec.Command("pppd", args...)
	pppdIn, err := pppdCmd.StdinPipe()
	if err != nil {
		return err
	}
	pppdCmd.Stdout = p.unescaper
	err = pppdCmd.Start()
	if err != nil {
		return err
	}
	p.commandInst = pppdCmd
	p.stdin = pppdIn
	// TODO: pipe stderr to log?

	p.IsStarted = true

	// TODO: remove this?
	go func() {
		defer log.Print("pppd disconnected")
		pppdCmd.Wait()
	}()

	return nil
}

// Kill kills pppd
func (p PppdInstance) Kill() error {
	if p.IsStarted && p.commandInst != nil {
		err := p.commandInst.Process.Kill()
		if err != nil {
			return err
		}
	}
	p.IsStarted = false
	return nil
}

func (p PppdInstance) Write(data []byte) (int, error) {
	return p.stdin.Write(pppEscape(data))
}

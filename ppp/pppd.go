package ppp

import (
	"io"
	"log"
	"os/exec"
)

// pppdConnection represents a connection to pppd
type pppdConnection struct {
	Config
	commandInst *exec.Cmd
	stdin       io.WriteCloser
	unescaper   pppUnescaper
	isStarted   bool
}

// start starts pppd
func (p *pppdConnection) start() error {
	p.unescaper = newUnescaper(p.DestWriter)

	// TODO: parse IP data
	args := append([]string{"notty", "file", "/etc/ppp/options.sstpd"}, p.ExtraArguments...)
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

	p.isStarted = true

	// TODO: remove this?
	go func() {
		defer log.Print("pppd disconnected")
		pppdCmd.Wait()
	}()

	return nil
}

// Close kills pppd if it is still running
func (p *pppdConnection) Close() error {
	if p.isStarted && p.commandInst != nil {
		err := p.commandInst.Process.Kill()
		if err != nil {
			return err
		}
	}
	p.isStarted = false
	return nil
}

func (p *pppdConnection) Write(data []byte) (int, error) {
	// TODO: check it's still running?
	return p.stdin.Write(pppEscape(data))
}

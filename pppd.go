package sstp

import (
	"io"
	"log"
	"net"
	"os/exec"
)

type pppdInstance struct {
	commandInst *exec.Cmd
	stdin       io.WriteCloser
	unescaper   pppUnescaper
	args        []string
}

type packetHandler struct {
	conn     net.Conn
	packChan chan []byte
}

func (p packetHandler) Write(data []byte) (int, error) {
	packetBytes := packDataPacketFast(data)
	p.packChan <- packetBytes
	return len(data), nil
}

func createPPPD(pppdInstance *pppdInstance, conn net.Conn) {
	args := append([]string{"notty", "file", "/etc/ppp/options.sstpd"}, pppdInstance.args...)
	pppdCmd := exec.Command("pppd", args...)
	pppdIn, err := pppdCmd.StdinPipe()
	handleErr(err)
	pppdCmd.Stdout = pppdInstance.unescaper
	err = pppdCmd.Start()
	handleErr(err)
	pppdInstance.commandInst = pppdCmd
	pppdInstance.stdin = pppdIn

	go func() {
		defer log.Print("pppd disconnected")
		pppdCmd.Wait()
	}()
}

package sstp

import (
	"fmt"
	"net"
	"net/http"

	"github.com/mholt/caddy"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Server is a httpserver.Handler that handles SSTP requests
type Server struct {
	httpTransport http.Transport
	NextHandler   httpserver.Handler
	testArg       string
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == "SSTP_DUPLEX_POST" {
		fmt.Print("Got a sstp request")
		return 200, nil
	}
	fmt.Print("Got a request")
	return s.NextHandler.ServeHTTP(w, r)
}

// Listener is a wrapper around a caddy.Listener that modifies SSTP requests
type Listener struct {
	caddy.Listener
}

// Accept is a wrapper around caddy.Listener.Accept() that intercepts the SSTP HTTP handshake
func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// TODO implement modifier
	addr, ok := c.RemoteAddr().(*net.TCPAddr)
	_ = addr
	if !ok {
		return c, nil
	}
	return c, nil
}

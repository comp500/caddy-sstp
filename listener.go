package sstp

import (
	"net"

	"github.com/mholt/caddy"
)

// The SSTP handshake uses a Content-Length greater than an int64, so it must be modified to be compatible

// Listener is a wrapper around a caddy.Listener that modifies SSTP requests
type Listener struct {
	caddy.Listener
	capturedHandshake string
}

// WrappedConn is a wrapper around a net.Conn that modifies SSTP requests
type WrappedConn struct {
	net.Conn
}

// Accept is a wrapper around caddy.Listener.Accept() that intercepts the SSTP HTTP handshake
func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return WrappedConn{Conn: c}, nil
}

// Overrides net.Conn.Read to modify SSTP requests
func (c WrappedConn) Read(b []byte) (n int, err error) {
	return c.Conn.Read(b)
}

package sstp

import (
	"bytes"
	"errors"
	"log"
	"net"

	"github.com/mholt/caddy"
)

// Listener is a wrapper around a caddy.Listener that modifies SSTP requests.
type Listener struct {
	caddy.Listener
}

// WrappedConn is a wrapper around a net.Conn that modifies SSTP requests.
type WrappedConn struct {
	net.Conn
	ignoreFurther   bool
	checkedMethod   bool
	currentOffset   int
	handshakeBuffer bytes.Buffer
}

// Accept is a wrapper around caddy.Listener.Accept() that intercepts the SSTP HTTP handshake.
func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &WrappedConn{Conn: c}, nil
}

// The string to find and replace within SSTP handshakes.
const (
	HandshakeLengthOriginal = "Content-Length: 18446744073709551615"
	HandshakeLengthReplaced = "Content-Length: 9223372036854775807 "
)

// The string to verify from the start of the request, to check that it is a SSTP handshake.
const (
	MethodCheckString = "SSTP"
	MethodCheckLen    = len(MethodCheckString)
)

// Overrides net.Conn.Read to modify SSTP requests.
//
// This is needed as SSTP handshakes use a Content-Length greater than an int64, so it must be modified to be compatible.
//
// This currently only works on HTTP requests, as we currently have no way intercept after SSL decryption.
//
// This function is passive: it will not read more bytes than c.Conn.Read reads, and will modify them if it is needed.
func (c *WrappedConn) Read(b []byte) (int, error) {
	// TODO: use buffer pools
	if c.ignoreFurther {
		return c.Conn.Read(b)
	}

	n, err := c.Conn.Read(b)
	if err != nil {
		return n, err
	}
	log.Printf("Read %v bytes", n)

	// Check the method
	if !c.checkedMethod {
		if c.currentOffset == 0 {
			if len(b) >= MethodCheckLen {
				// If not SSTP, ignore further bytes and passthrough
				if !bytes.Equal(b[:MethodCheckLen], []byte(MethodCheckString)) {
					c.ignoreFurther = true
					return n, nil
				}
				c.checkedMethod = true
			} else {
				// Return with current data, wait for more
				c.handshakeBuffer.Write(b)
				c.currentOffset += len(b)
				return n, nil
			}
		} else {
			c.handshakeBuffer.Write(b)
			if c.handshakeBuffer.Len() < MethodCheckLen {
				// Return with current data, wait for more
				c.currentOffset += c.handshakeBuffer.Len()
				return n, nil
			}
			checkSlice := make([]byte, MethodCheckLen)
			numChecked, err := c.handshakeBuffer.Read(checkSlice)
			if err != nil {
				return n, err
			}
			if numChecked != MethodCheckLen {
				// This shouldn't happen
				return n, errors.New("Internal sstp WrappedConn error")
			}

			// If not SSTP, ignore further bytes and passthrough
			if !bytes.Equal(checkSlice, []byte(MethodCheckString)) {
				c.ignoreFurther = true
				return n, nil
			}
			c.checkedMethod = true
		}
	}

	log.Print("SSTP packet received!")

	// TODO: fix for multiple Read() calls, if Content-Length is split across multiple
	// Find Content-Length
	index := bytes.Index(b, []byte(HandshakeLengthOriginal))
	if index > -1 {
		// replace "Content-Length: 18446744073709551615"
		// with    "Content-Length: 9223372036854775807 " (max int64)
		replaced := []byte(HandshakeLengthReplaced)
		for i := 0; i < 36; i++ {
			b[index+i] = replaced[i]
		}
		c.ignoreFurther = true
	} else {
		log.Print("Content-Length not found")
	}

	return n, nil
}

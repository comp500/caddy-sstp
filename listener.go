package sstp

import (
	"bytes"
	"errors"
	"fmt"
	"net"

	"github.com/mholt/caddy"
)

// The SSTP handshake uses a Content-Length greater than an int64, so it must be modified to be compatible

// Listener is a wrapper around a caddy.Listener that modifies SSTP requests
type Listener struct {
	caddy.Listener
}

// WrappedConn is a wrapper around a net.Conn that modifies SSTP requests
type WrappedConn struct {
	net.Conn
	capturedHandshake string
	ignoreFurther     bool
	checkedMethod     bool
	currentOffset     int
	handshakeBuffer   bytes.Buffer
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
// TODO: use buffer pools
// This currently only works on HTTP requests, as we currently have no way intercept after SSL decryption
func (c WrappedConn) Read(b []byte) (int, error) {
	if c.ignoreFurther {
		return c.Conn.Read(b)
	}

	n, err := c.Conn.Read(b)
	if err != nil {
		return n, err
	}
	fmt.Printf("Read %v bytes\n", n)

	// Check the method
	if !c.checkedMethod {
		if c.currentOffset == 0 {
			if len(b) >= 4 {
				// If not SSTP, ignore further bytes and passthrough
				if !bytes.Equal(b[:4], []byte("SSTP")) {
					c.ignoreFurther = true
					return n, nil
				}
				c.checkedMethod = true
			}
		} else {
			c.handshakeBuffer.Write(b)
			if c.handshakeBuffer.Len() < 4 {
				// Return with current data, wait for more
				c.currentOffset += c.handshakeBuffer.Len()
				return n, nil
			}
			checkSlice := make([]byte, 4)
			numChecked, err := c.handshakeBuffer.Read(checkSlice)
			if err != nil {
				return n, err
			}
			if numChecked != 4 {
				// This shouldn't happen
				return n, errors.New("Internal sstp WrappedConn error")
			}

			// If not SSTP, ignore further bytes and passthrough
			if !bytes.Equal(checkSlice, []byte("SSTP")) {
				c.ignoreFurther = true
				return n, nil
			}
			c.checkedMethod = true
		}
	}

	fmt.Print("SSTP packet received!")

	// TODO: fix for multiple Read() calls, if Content-Length is split across multiple
	// Find Content-Length
	index := bytes.Index(b, []byte("Content-Length: 18446744073709551615"))
	if index > -1 {
		// replace "Content-Length: 18446744073709551615"
		// with    "Content-Length: 9223372036854775807 " (max uint64)
		replaced := []byte("Content-Length: 9223372036854775807 ")
		for i := 0; i < 36; i++ {
			b[index+i] = replaced[i]
		}
		c.ignoreFurther = true
	} else {
		fmt.Print("Content-Length not found")
	}

	return n, err
}

package ppp

import (
	"errors"
	"fmt"
	"io"
	"net"
)

// Config defines the settings that the PPP connection should use
type Config struct {
	DestIP         net.IP
	SrcIP          net.IP
	ExtraArguments []string
	ConnectionType ConnectionType
	DestWriter     io.Writer
}

// ConnectionType is the connection method used by a connection
type ConnectionType int

// Constants for ConnectionType values
const (
	ConnectionTypeNative = 0
	ConnectionTypePppd   = 1
)

func (k ConnectionType) String() string {
	switch k {
	case ConnectionTypeNative:
		return "Native"
	case ConnectionTypePppd:
		return "Pppd"
	default:
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

// Connection is a PPP connection instance
type Connection interface {
	io.WriteCloser
	start() error
}

// NewConnection starts a new PPP connection from the given config
func NewConnection(config Config) (*Connection, error) {
	var conn Connection

	switch config.ConnectionType {
	case ConnectionTypeNative:
		conn = &nativeConnection{Config: config}
	case ConnectionTypePppd:
		conn = &PppdInstance{Config: config}
	default:
		return nil, errors.New("Connection type not supported")
	}

	err := conn.start()
	if err != nil {
		return nil, err
	}

	return &conn, nil
}

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
	ConnectionTypeTunTap ConnectionType = iota
	ConnectionTypeVirtualNAT
	ConnectionTypePppd
)

func (k ConnectionType) String() string {
	switch k {
	case ConnectionTypeTunTap:
		return "TUN/TAP"
	case ConnectionTypeVirtualNAT:
		return "VirtualNAT"
	case ConnectionTypePppd:
		return "pppd"
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
	case ConnectionTypeTunTap:
		conn = &nativeConnection{Config: config, Vnat: false}
	case ConnectionTypeVirtualNAT:
		conn = &nativeConnection{Config: config, Vnat: true}
	case ConnectionTypePppd:
		conn = &pppdConnection{Config: config}
	default:
		return nil, errors.New("Connection type not supported")
	}

	err := conn.start()
	if err != nil {
		return nil, err
	}

	return &conn, nil
}

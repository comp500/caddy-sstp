package ppp

// This file manages pppd connections for the native (pure Go) connection type.
type nativeConnection struct {
	Config
}

func (p nativeConnection) Write(data []byte) (int, error) {
	return 0, nil
}

func (p nativeConnection) Close() error {
	return nil
}

func (p nativeConnection) start() error {
	return nil
}

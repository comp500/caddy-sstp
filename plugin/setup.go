// Package plugin is a Caddy HTTP plugin that implements the SSTP protocol.
// This requires pppd to bridge the connection through SSTP.
package plugin

import (
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("sstp", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	server := &Server{}

	for c.Next() { // skip the directive name
		server.pppdArgs = c.RemainingArgs()
	}

	cfg := httpserver.GetConfig(c)
	mid := func(next httpserver.Handler) httpserver.Handler {
		server.NextHandler = next
		return server
	}
	cfg.AddMiddleware(mid)
	listenMid := func(next caddy.Listener) caddy.Listener {
		listen := &Listener{Listener: next}
		return listen
	}
	cfg.AddListenerMiddleware(listenMid)

	return nil
}

// Package plugin is a Caddy HTTP plugin that implements the SSTP protocol.
// This requires pppd to bridge the connection through SSTP.
package plugin

import (
	"net"

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
		for c.NextBlock() {
			directive := c.Val()
			args := c.RemainingArgs()
			switch directive {
			case "args":
				server.extraArgs = args
			case "src_ip":
				if len(args) != 1 {
					return c.ArgErr()
				}
				server.srcIP = net.ParseIP(args[0])
				if server.srcIP == nil { // parsing failed
					return c.ArgErr()
				}
			case "dest_ip":
				if len(args) != 1 {
					return c.ArgErr()
				}
				server.destIP = net.ParseIP(args[0])
				if server.destIP == nil { // parsing failed
					return c.ArgErr()
				}
			default:
				return c.ArgErr()
			}
		}
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

package sstp

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
		if !c.NextArg() { // expect at least one value
			return c.ArgErr() // otherwise it's an error
		}
		server.testArg = c.Val() // use the value
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

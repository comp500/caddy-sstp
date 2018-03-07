package sstp

import (
	"log"
	"net/http"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Server is a httpserver.Handler that handles SSTP requests
type Server struct {
	httpTransport http.Transport
	Next          httpserver.Handler
	testArg       string
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	log.Print("Got a request")
	return s.Next.ServeHTTP(w, r)
}

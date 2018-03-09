package sstp

import (
	"fmt"
	"net/http"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Server is a httpserver.Handler that handles SSTP requests.
type Server struct {
	NextHandler httpserver.Handler
	testArg     string
}

// Serves SSTP requests. See httpserver.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == "SSTP_DUPLEX_POST" {
		fmt.Print("Got a sstp request")
		return 200, nil
	}
	fmt.Print("Got a request")
	return s.NextHandler.ServeHTTP(w, r)
}

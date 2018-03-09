package sstp

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Server is a httpserver.Handler that handles SSTP requests.
type Server struct {
	NextHandler httpserver.Handler
	testArg     string
}

// MethodSstp is the SSTP handshake's HTTP method
const MethodSstp = "SSTP_DUPLEX_POST"

// Serves SSTP requests. See httpserver.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == MethodSstp {
		// TODO: check URI
		if r.ProtoMajor != 1 { // We don't support HTTP2
			return http.StatusHTTPVersionNotSupported, errors.New("Unsupported HTTP major version: " + strconv.Itoa(r.ProtoMajor))
		}

		fmt.Print("Got a sstp request")
		return 200, nil
	}
	fmt.Print("Got a request")
	return s.NextHandler.ServeHTTP(w, r)
}

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

// MethodSstp is the SSTP handshake's HTTP method.
const MethodSstp = "SSTP_DUPLEX_POST"

// RequestPath is the path that the SSTP handshake uses.
const RequestPath = "/sra_{BA195980-CD49-458b-9E23-C84EE0ADCD75}/"

// Serves SSTP requests. See httpserver.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == MethodSstp {
		if httpserver.Path(r.URL.Path).Matches(RequestPath) {
			if r.ProtoMajor != 1 { // We don't support HTTP2
				return http.StatusHTTPVersionNotSupported, errors.New("Unsupported HTTP major version: " + strconv.Itoa(r.ProtoMajor))
			}

			fmt.Print("Got a sstp request")

			hijacker, ok := w.(http.Hijacker)
			if !ok {
				return http.StatusInternalServerError, errors.New("ResponseWriter does not implement Hijacker")
			}
			// Hijack connection
			// Ignore returned bufio.Reader, as client will only give more data after first response
			clientConn, _, err := hijacker.Hijack()
			if err != nil {
				return http.StatusInternalServerError, errors.New("failed to hijack: " + err.Error())
			}
			defer clientConn.Close()

			n, err := fmt.Fprintf(clientConn, "%s\r\n%s\r\n%s\r\n%s\r\n\r\n",
				"HTTP/1.1 200 OK",
				"Date: Thu, 09 Nov 2006 00:51:09 GMT",
				"Server: Microsoft-HTTPAPI/2.0",
				"Content-Length: 18446744073709551615")
			if err != nil {
				return http.StatusInternalServerError, errors.New("Failed to send response")
			}

			fmt.Printf("Written %v bytes http response", n)

			return 0, nil
		}
	}
	fmt.Print("Got a request")
	return s.NextHandler.ServeHTTP(w, r)
}

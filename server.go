package sstp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
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

// TODO: don't use this horrible function
func handleErr(err error) {
	if err != nil {
		log.Fatalf("%s\n", err)
	}
}

// Serves SSTP requests. See httpserver.Handler.
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == MethodSstp {
		if httpserver.Path(r.URL.Path).Matches(RequestPath) {
			if r.ProtoMajor != 1 { // We don't support HTTP2
				return http.StatusHTTPVersionNotSupported, errors.New("Unsupported HTTP major version: " + strconv.Itoa(r.ProtoMajor))
			}

			log.Print("Got a sstp request")

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

			n, err := fmt.Fprintf(clientConn, "%s\r\n%s\r\n%s\r\n%s\r\n\r\n",
				"HTTP/1.1 200 OK",
				"Date: Thu, 09 Nov 2006 00:51:09 GMT",
				"Server: Microsoft-HTTPAPI/2.0",
				"Content-Length: 18446744073709551615")
			if err != nil {
				clientConn.Close()
				return http.StatusInternalServerError, errors.New("Failed to send response")
			}

			log.Printf("Written %v bytes http response", n)

			// Pass the connection to handleConnection
			go s.handleConnection(clientConn)

			return 0, nil
		}
	}
	log.Print("Got a request")
	return s.NextHandler.ServeHTTP(w, r)
}

type parseReturn struct {
	isControl bool
	Data      []byte
}

// Handles SSTP connection after HTTP 200 is sent
func (s Server) handleConnection(c net.Conn) {
	// Shut down the connection.
	defer c.Close()

	// TODO: make more idiomatic, just copied straight from sstp-go

	ch := make(chan parseReturn)
	eCh := make(chan error)

	packChan := make(chan []byte)
	pppdInstance := pppdInstance{nil, nil, newUnescaper(packetHandler{c, packChan})} // store null pointer to future pppd instance

	// Start a goroutine to read from our net connection
	go func(ch chan parseReturn, eCh chan error) {
		for {
			// try to read the data
			var data [4]byte
			n, err := c.Read(data[:])
			if err != nil {
				// send an error if it's encountered
				eCh <- err
				return
			}
			if n < 4 {
				eCh <- errors.New("Invalid packet")
				return
			}
			isControl, lengthToRead, err := decodeHeader(data[:])
			if err != nil {
				// send an error if it's encountered
				eCh <- err
				return
			}
			newData := make([]byte, lengthToRead)
			n, err = c.Read(newData)
			if err != nil {
				// send an error if it's encountered
				eCh <- err
				return
			}
			if n != lengthToRead {
				eCh <- errors.New("Not all of packet read")
				return
			}
			ch <- parseReturn{isControl, newData}
		}
	}(ch, eCh)

	go func(packChan chan []byte) {
		for {
			select {
			case data := <-packChan: // This case means we recieved data on the connection
				c.Write(data)
			}
		}
	}(packChan)

	//ticker := time.Tick(time.Second)
	// continuously read from the connection
	for {
		select {
		case data := <-ch: // This case means we recieved data on the connection
			// Do something with the data
			//log.Printf("%s\n", hex.Dump(data))
			if data.isControl {
				header := parseControl(data.Data)
				handleControlPacket(header, c, &pppdInstance)
			} else {
				handleDataPacket(data.Data, c, &pppdInstance)
			}
		case err := <-eCh: // This case means we got an error and the goroutine has finished
			if err == io.EOF {
				log.Print("Client disconnected")
				if pppdInstance.commandInst != nil {
					// kill pppd if disconnect
					err := pppdInstance.commandInst.Process.Kill()
					handleErr(err)
					pppdInstance.commandInst = nil
				}
			} else {
				log.Fatalf("%s\n", err)
				// handle our error then exit for loop
				break
				// This will timeout on the read.
				//case <-ticker:
				// do nothing? this is just so we can time out if we need to.
				// you probably don't even need to have this here unless you want
				// do something specifically on the timeout.
			}
		}
	}
}

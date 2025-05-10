package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
)

var (
	// ErrResponseWriterNoHijack is returned when the ResponseWriter does not implement the Hijacker interface.
	ErrResponseWriterNoHijack = errors.New("response writer cannot be hijacked: does not implement Hijacker interface")
)
var (
	// ResponseHijackingError is the response sent to the client when the proxy cannot hijack the connection.
	ResponseHijackingError []byte = []byte("hijacking error")
	// ResponseRawSuccess is the response sent to the client when the HTTPS proxy successfully hijacks the connection.
	ResponseRawSuccess []byte = []byte("HTTP/1.1 200 OK\r\n" +
		"Connection: close\r\n" +
		"Content-Length: 0\r\n" +
		"\r\n")
)

// look over HTTP/2 handling and streaming responses

// handleHTTP handles a HTTP request to the proxy.
//
// HTTP requests do not create secure tunnels, therefore the original request to the
// proxy is the one meant for the host. The only prep we need to do is strip proxy
// headers. The proxy will then perform the request to the host and send the response
// back to the client.
func handleHTTP(config *Config, conn net.Conn, r *http.Request) error {
	// perform the request
	resp, err := perform(config, r, false)
	if err != nil {
		return fmt.Errorf("perform: %w", err)
	}

	_, err = conn.Write(resp)
	return err
}

// handleHTTPS handles a HTTPS request. This proxy is a MITM proxy, so it will
// use a self-signed certificate for the requested host and send it to the client.
//
// Steps:
// (1) It writes a success response to the client. The client believes it is no longer speaking to the proxy now.
// (2) A TLS handshake (using the self-signed) occurs between the client and the proxy. The client believes it is speaking to the host.
// (3) The proxy reads the actual request from the client that it meant to send the host.
// (4) The proxy performs the request to the real host.
// (5) The proxy sends the response back to the client.
func handleHTTPS(config *Config, conn net.Conn, c *Certificates, r *http.Request) error {
	_, err := conn.Write(ResponseRawSuccess)
	if err != nil {
		return fmt.Errorf("connection write: %w", err)
	}

	// after the success response, a handshake will occur and the user will
	// send the ACTUAL request

	// NOTE: consider IPV6 square brackets and support for it
	tlsconn, err := c.TLSConn(config, conn, r.URL.Hostname())
	if err != nil {
		return fmt.Errorf("tls conn: %w", err)
	}

	finalReq, err := http.ReadRequest(bufio.NewReader(tlsconn))
	if err != nil {
		return fmt.Errorf("read mitm request: %w", err)
	}

	resp, err := perform(config, finalReq, true)
	if err != nil {
		return fmt.Errorf("perform: %w", err)
	}

	_, err = tlsconn.Write(resp)
	return err
}

// hijack hijacks the ResponseWriter connection to the client.
func hijack(w http.ResponseWriter) (net.Conn, error) {
	h, ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrResponseWriterNoHijack
	}
	c, _, err := h.Hijack()
	return c, err
}

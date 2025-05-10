package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
)

var (
	ErrResponseWriterNoHijack = errors.New("response writer cannot be hijacked: does not implement Hijacker interface")
)
var (
	ResponseHijackingError []byte = []byte("hijacking error")
	ResponseRawSuccess     []byte = []byte("HTTP/1.1 200 OK\r\n" +
		"Connection: close\r\n" +
		"Content-Length: 0\r\n" +
		"\r\n")
)

// look over HTTP/2 handling and streaming responses

// handleHTTP handles a HTTP request to the proxy.
//
// HTTP request:
// client (regular http request + proxy headers) -> proxy (sending the client's request stripped of proxy headers) -> host
// and then backwards
func handleHTTP(config *Config, conn net.Conn, r *http.Request) error {
	// perform the request
	resp, err := perform(config, r, false)
	if err != nil {
		return fmt.Errorf("perform: %w", err)
	}

	_, err = conn.Write(resp)
	return err
}

// handleHTTPS handles a HTTPS request.
//
// client (request to proxy, includes host) -> proxy -> host
// proxy tunnels this but can't see any information like path
//
// THIS proxy however will use MITM to intercept the request
func handleHTTPS(config *Config, conn net.Conn, c *Certificates, r *http.Request) error {
	_, err := conn.Write(ResponseRawSuccess)
	if err != nil {
		return fmt.Errorf("connection write: %w", err)
	}

	// after the success response, a handshake will occur and the user will
	// send the ACTUAL request

	// NOTE: consider IPV6 square brackets and support for it
	tlsconn, err := c.TLSConn(conn, r.URL.Hostname())
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

func hijack(w http.ResponseWriter) (net.Conn, error) {
	h, ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrResponseWriterNoHijack
	}
	c, _, err := h.Hijack()
	return c, err
}

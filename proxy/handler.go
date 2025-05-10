package main

import (
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
)

// look over HTTP/2 handling and streaming responses

func handleHTTP(config *Config, conn net.Conn, r *http.Request) error {
	// perform the request
	resp, err := perform(config, r)
	if err != nil {
		return fmt.Errorf("perform: %w", err)
	}

	_, err = conn.Write(resp)
	return err
}

func handleHTTPS() error {
	panic("not handling this rn :/")
	return nil
}

func hijack(w http.ResponseWriter) (net.Conn, error) {
	h, ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrResponseWriterNoHijack
	}
	c, _, err := h.Hijack()
	return c, err
}

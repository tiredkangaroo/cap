package main

import (
	"errors"
	"log/slog"
	"net"

	"github.com/tiredkangaroo/bigproxy/proxy/http"

	certificate "github.com/tiredkangaroo/bigproxy/proxy/certificates"
	"github.com/tiredkangaroo/bigproxy/proxy/config"
)

// ProxyHandler is being used here because CONNECT requests are automatically
// rejected by net/http's default handler on ListenAndServe.
type ProxyHandler struct {
	certifcates *certificate.Certificates
	m           *Manager
}

func (c *ProxyHandler) ServeHTTP(conn net.Conn, r *http.Request) {
	defer conn.Close()

	req := new(Request)
	req.Init(conn, r)

	c.serveAfterInit(req, r)
}

func (c *ProxyHandler) serveAfterInit(req *Request, r *http.Request) {
	c.m.SendNew(req)

	var err error
	if req.Secure { // we're handling an HTTPS connection here
		err = req.handleHTTPS(c.m, c.certifcates)
	} else {
		// we're handling an HTTP connection here, HTTP connections send the full request to the proxy, so we actually do need this request here
		err = req.handleHTTP(c.m, r)
	}

	if errors.Is(err, ErrPerformStop) { // ignore PerformStop errors and don't send a control message
		return
	}

	if err != nil {
		slog.Error("request handling", "err", err.Error())
		c.m.SendError(req, err)
	} else {
		c.m.SendDone(req)
	}
}

func (c *ProxyHandler) Init(dirname string) error {
	c.certifcates = new(certificate.Certificates)
	if err := c.certifcates.Init(dirname); err != nil {
		slog.Warn("initializing certificates (mitm cannot be used)", "err", err.Error())
		c.certifcates = nil
		config.DefaultConfig.MITM = false
	}
	return nil
}

func (c *ProxyHandler) ListenAndServe(m *Manager, dirname string) error {
	// HTTP pathway:
	// read request -> init request -> dial host -> write request -> read response -> write response
	// HTTPS NO MITM pathway:
	// read request -> init request -> send 200 response -> tunnel
	// HTTPS MITM pathway:
	// read request -> init request -> send 200 response -> cert gen + handshake -> dial host -> write request -> read response -> write response

	if err := c.Init(dirname); err != nil {
		return err
	}
	c.m = m

	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		slog.Error("failed to start proxy listener", "err", err.Error())
		return err
	}
	for {
		rawconn, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection", "err", err.Error())
			continue
		}
		conn := NewCustomConn(rawconn)

		req, err := http.ReadRequest(conn)
		if err != nil {
			slog.Error("failed to read request", "err", err.Error())
			conn.Close()
			continue
		}

		go func() {
			defer req.Body.Close()
			c.ServeHTTP(conn, req)
		}()
	}
}

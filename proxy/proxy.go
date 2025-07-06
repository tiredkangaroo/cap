package main

import (
	"errors"
	"log/slog"
	"net"

	"github.com/google/uuid"
	"github.com/tiredkangaroo/cap/proxy/http"
	"github.com/tiredkangaroo/cap/proxy/timing"

	certificate "github.com/tiredkangaroo/cap/proxy/certificates"
	"github.com/tiredkangaroo/cap/proxy/config"
)

// ProxyHandler is being used here because CONNECT requests are automatically
// rejected by net/http's default handler on ListenAndServe.
type ProxyHandler struct {
	certifcates *certificate.Certificates
	m           *Manager
	tcaservice  *TCAService
}

func (c *ProxyHandler) ServeHTTP(pr *Request, r *http.Request) {
	pr.Init(r)

	c.serveAfterInit(pr, r)
}

func (c *ProxyHandler) serveAfterInit(req *Request, r *http.Request) {
	c.m.SendNew(req)

	var err error
	if req.Secure { // we're handling an HTTPS connection here
		err = req.handleHTTPS(c.m, c.certifcates)
	} else {
		// we're handling an HTTP connection here, HTTP connections send the full request to the proxy, so we actually do need this request here
		err = req.handleHTTP(c.m, r, c.certifcates)
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
		r := new(Request)
		t := timing.New(m.setStateFunc(r))
		conn := NewCustomConn(rawconn)

		r.conn = conn
		id, err := uuid.NewRandom()
		if err != nil {
			slog.Error("uuid error", "err", err.Error())
			r.ID = "75756964-7634-6765-6e65-72726f720000" // this isn't a random UUID
		}
		r.ID = id.String()
		r.timing = t

		t.Start(timing.TimeReadProxyRequest)
		req, err := http.ReadRequest(conn)
		if err != nil {
			slog.Error("failed to read request", "err", err.Error())
			conn.Close()
			continue
		}
		t.Stop()

		go func() {
			defer req.Body.CloseBody()
			defer conn.Close()
			c.ServeHTTP(r, req)
		}()
	}
}

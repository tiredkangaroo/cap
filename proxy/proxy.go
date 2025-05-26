package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	certificate "github.com/tiredkangaroo/bigproxy/proxy/certificates"
	"github.com/tiredkangaroo/bigproxy/proxy/config"
)

// ProxyHandler is being used here because CONNECT requests are automatically
// rejected by net/http's default handler on ListenAndServe.
type ProxyHandler struct {
	certifcates *certificate.Certificates

	// liveRequestMessages messages:
	//
	// NEW {id: string, secure: bool, clientIP: string, clientAuthorization: string, host: string}
	// - followed up by: "HTTP {id: string, method: string, headers: map[string][]string, body: []byte}"
	// - followed up by: "HTTPS-MITM-REQUEST {id: string, method: string, headers: map[string][]string, body: []byte}"
	// - followed up by: "HTTPS-TUNNEL-REQUEST "
	// - followed up by: "HTTP-RESPONSE {id: string, statusCode: int, headers: map[string][]string, body: []byte}"
	// - followed up by: "HTTPS-MITM-RESPONSE {id: string, statusCode: int, headers: map[string][]string, body: []byte}"
	// - followed up by: "HTTPS-TUNNEL-RESPONSE "
	m *Manager
}

func (c *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// c.working.Add(1)
	// defer c.working.Add(-1)

	req := new(Request)
	req.Init(w, r)
	defer req.conn.Close()

	c.m.SendNew(req)

	var err error
	if req.secure { // we're handling an HTTPS connection here
		err = req.handleHTTPS(c.m, c.certifcates)
	} else {
		// we're handling an HTTP connection here
		err = req.handleHTTP(c.m)
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

func (c *ProxyHandler) Init() error {
	c.certifcates = new(certificate.Certificates)
	if err := c.certifcates.Init(); err != nil {
		return err
	}
	return nil
}

func (c *ProxyHandler) ListenAndServe(m *Manager) {
	ph := new(ProxyHandler)
	ph.certifcates = new(certificate.Certificates)
	ph.m = m

	if err := ph.certifcates.Init(); err != nil {
		slog.Error("initializing certificates", "err", err.Error())
		config.DefaultConfig.MITM = false // NOTE: add errors when trying to set MITM true when initialization fails
	}

	if err := http.ListenAndServe(":8000", ph); err != nil {
		slog.Error("fatal proxy server", "err", err.Error())
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

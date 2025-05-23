package main

import (
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
	controlMessages chan []byte
}

func (c *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// c.working.Add(1)
	// defer c.working.Add(-1)

	req := new(Request)
	req.Init(w, r)
	defer req.conn.Close()

	sendControlNew(req, c.controlMessages)

	var err error
	if req.secure { // we're handling an HTTPS connection here
		err = req.handleHTTPS(c.certifcates, c.controlMessages)
	} else {
		// we're handling an HTTP connection here
		err = req.handleHTTP(c.controlMessages)
	}

	if err != nil {
		slog.Error("request handling", "err", err.Error())
	}
}

func (c *ProxyHandler) Init() error {
	c.certifcates = new(certificate.Certificates)
	if err := c.certifcates.Init(); err != nil {
		return err
	}
	return nil
}

func (c *ProxyHandler) ListenAndServe(controlMessages chan []byte) {
	ph := new(ProxyHandler)
	ph.certifcates = new(certificate.Certificates)
	ph.controlMessages = controlMessages

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

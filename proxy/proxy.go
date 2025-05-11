package main

import (
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

const window = time.Second * 20

// ProxyHandler is being used here because CONNECT requests are automatically
// rejected by net/http's default handler on ListenAndServe.
type ProxyHandler struct {
	config           *Config
	working          atomic.Int32
	requestsInWindow atomic.Uint32
	avgMSPerRequest  float64

	certifcates *Certificates

	// fc is a score determining for priority of requests.
	fc float64
}

func (c *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.working.Add(1)
	defer c.working.Add(-1)

	conn, err := hijack(w)
	if err != nil {
		w.WriteHeader(400)
		w.Write(ResponseHijackingError)
		return
	}
	defer conn.Close()

	if r.Method == http.MethodConnect { // we're handling an HTTPS connection here
		err = handleHTTPS(c.config, conn, c.certifcates, r)
	} else {
		// we're handling an HTTP connection here
		err = handleHTTP(c.config, conn, r)
	}

	if err != nil {
		slog.Error("request handling", "err", err.Error())
	}
}

func main() {
	config := new(Config)
	go startControlServer(config)

	// -log([H+]) im so funny
	ph := new(ProxyHandler)
	ph.config = config
	ph.certifcates = new(Certificates)
	if err := ph.certifcates.Init(); err != nil {
		slog.Error("fatal certificates init", "err", err.Error())
	}

	if err := http.ListenAndServe(":8000", ph); err != nil {
		slog.Error("fatal proxy server", "err", err.Error())
	}
}

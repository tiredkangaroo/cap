package main

import (
	"log/slog"
	"sync"

	_ "github.com/tiredkangaroo/bigproxy/proxy/config"
	"github.com/tiredkangaroo/websocket"
)

func main() {
	m := &Manager{
		wsConns:             make([]*websocket.Conn, 0, 8),
		approvalWaiters:     make(map[string]*Request, 8),
		approvalWaitersRWMu: sync.RWMutex{},
	}

	go startControlServer(m)

	ph := new(ProxyHandler)
	if err := ph.Init(); err != nil {
		slog.Error("initializing proxy handler", "err", err.Error())
		return
	}

	ph.ListenAndServe(m)
}

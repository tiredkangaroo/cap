package main

import (
	"log/slog"

	_ "github.com/tiredkangaroo/bigproxy/proxy/config"
)

func main() {
	controlMessages := &ControlChannel{
		u:                       make(chan []byte, 8),
		waitingApprovalResponse: make(map[string]*Request, 8),
	}

	go startControlServer(controlMessages)

	ph := new(ProxyHandler)
	if err := ph.Init(); err != nil {
		slog.Error("initializing proxy handler", "err", err.Error())
		return
	}

	ph.ListenAndServe(controlMessages)
}

package main

import (
	"log/slog"
	"net"
	"os"
	"sync"

	_ "github.com/tiredkangaroo/bigproxy/proxy/config"
	"github.com/tiredkangaroo/websocket"
)

var myLocalIP string
var myPID int

func init() {
	c, err := net.Dial("tcp", "1.1.1.1:443")
	if err == nil {
		myLocalIP, _, _ = net.SplitHostPort(c.LocalAddr().String())
	} else {
		slog.Error("failed to get local IP", "err", err.Error())
	}
	myPID = os.Getpid()
}

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

package main

import (
	_ "embed"

	"database/sql"
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
	sqldb, err := sql.Open("sqlite", "cap.db")
	if err != nil {
		slog.Error("failed to open database", "err", err.Error())
		return
	}
	db := NewDatabase(sqldb)
	if err := db.Init(); err != nil {
		slog.Error("failed to initialize database", "err", err.Error())
		return
	}

	m := &Manager{
		wsConns:             make([]*websocket.Conn, 0, 8),
		approvalWaiters:     make(map[string]*Request, 8),
		approvalWaitersRWMu: sync.RWMutex{},
		db:                  db,
	}

	go startControlServer(m)

	ph := new(ProxyHandler)
	ph.ListenAndServe(m)
}

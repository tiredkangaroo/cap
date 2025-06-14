package main

import (
	_ "embed"

	_ "github.com/tiredkangaroo/bigproxy/proxy/config"
	"github.com/tiredkangaroo/bigproxy/proxy/db"
	_ "modernc.org/sqlite"

	"database/sql"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/tiredkangaroo/websocket"
)

var myLocalIP string
var myPID int

//go:embed schema.sql
var ddl string

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
	if _, err := sqldb.Exec(ddl); err != nil {
		slog.Error("failed to execute schema", "err", err.Error())
		return
	}

	queries := db.New(sqldb)

	m := &Manager{
		wsConns:             make([]*websocket.Conn, 0, 8),
		approvalWaiters:     make(map[string]*Request, 8),
		approvalWaitersRWMu: sync.RWMutex{},
		queries:             queries,
	}

	go startControlServer(m)

	ph := new(ProxyHandler)
	ph.ListenAndServe(m)
}

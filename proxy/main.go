package main

import (
	_ "embed"

	"log/slog"
	"net"
	"os"

	_ "github.com/tiredkangaroo/bigproxy/proxy/config"
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
	db := NewDatabase()
	if err := db.Init(); err != nil {
		slog.Error("failed to initialize database", "err", err.Error())
		return
	}
	defer db.b.Close()

	m := NewManager(db)

	ph := new(ProxyHandler)
	go startControlServer(m, ph)
	ph.ListenAndServe(m)
}

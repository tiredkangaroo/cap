package main

import (
	_ "embed"
	"net/http"
	"net/http/pprof"
	"path/filepath"

	"log/slog"
	"net"
	"os"

	"github.com/google/uuid"
	"github.com/tiredkangaroo/bigproxy/proxy/config"
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

	uuid.DisableRandPool()
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)

	var dirname string
	if os.Getenv("DEBUG") != "true" {
		execFile, err := os.Executable()
		if err != nil {
			slog.Error("failed to get executable path", "err", err.Error())
			return
		}
		slog.Info("bigproxy running at", "dirname", filepath.Dir(execFile), "pid", myPID, "localIP", myLocalIP)
		dirname = filepath.Dir(execFile)
	} else {
		dirname = "."
	}

	db := NewDatabase()
	if err := db.Init(dirname); err != nil {
		slog.Error("failed to initialize database", "err", err.Error())
		return
	}
	defer db.b.Close()

	m := NewManager(db)

	// NOTE: ensure
	if config.DefaultConfig.Debug {
		mux := http.DefaultServeMux
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		slog.Info("debug mode enabled")
	} else {
		slog.Info("debug mode disabled, use DEBUG=true to enable it")
	}

	ph := new(ProxyHandler)
	go startControlServer(m, ph)
	ph.ListenAndServe(m, dirname)
}

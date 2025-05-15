package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
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
	config.Debug = os.Getenv("DEBUG") == "true"
	if config.Debug {
		slog.Info("debug mode")
	}
	// manageConfigFile is on the main thread to stop the program from terminating
	// before the signal handler catches the signal.
	manageConfigFile(config, os.Getenv("CONFIG_SAVEFILE"))

	go startControlServer(config)

	// -log([H+]) im so funny
	ph := new(ProxyHandler)
	ph.config = config
	ph.certifcates = new(Certificates)
	if err := ph.certifcates.Init(); err != nil {
		slog.Error("fatal certificates init", "err", err.Error())
	}

	if err := http.ListenAndServe(":8000", ph); err != nil {
		os.Exit(1)
		slog.Error("fatal proxy server", "err", err.Error())
	} else {
		os.Exit(0)
	}

}

func manageConfigFile(config *Config, filename string) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		slog.Error("creating/opening file specified in CONFIG_SAVEFILE", "err", err.Error())
		return
	}

	if err := readConfigFile(file, config); err != nil {
		slog.Error("reading config file", "err", err.Error())
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		defer os.Exit(0)
		<-c
		slog.Info("received signal, saving config file")
		data, err := json.Marshal(config)
		if err != nil {
			slog.Error("saving config file at marshal step", "err", err.Error())
			return
		}
		if err := file.Truncate(0); err != nil {
			slog.Error("saving config file at truncate step", "err", err.Error())
		}
		if _, err := file.Seek(0, 0); err != nil {
			slog.Error("saving config file at the seek 0 step", "err", err.Error())
		}
		if _, err := file.Write(data); err != nil {
			slog.Error("saving config file at file.Write step", "err", err.Error())
			return
		}
	}()

}

func readConfigFile(file *os.File, config *Config) error {
	rf, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rf, config)
	if err != nil {
		return err
	}

	return nil
}

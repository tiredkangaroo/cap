package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
	"github.com/tiredkangaroo/websocket"
)

func startControlServer(liveRequestMessages chan []byte) {
	liveRequestWebsockets := []*websocket.Conn{}
	// Start the control server
	http.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		if config.DefaultConfig.Debug {
			setCORSHeaders(w)
		}

		data, err := json.Marshal(config.DefaultConfig)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("failed to marshal config"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	http.HandleFunc("POST /config", func(w http.ResponseWriter, r *http.Request) {
		if config.DefaultConfig.Debug {
			setCORSHeaders(w)
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("failed to read request body"))
			return
		}

		// possible validation of the config here and authority to change it

		var newConfig config.Config
		err = json.Unmarshal(b, &newConfig)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("failed to decode config"))
			return
		}

		*config.DefaultConfig = newConfig
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("config updated"))
	})

	http.HandleFunc("GET /requestsWS", func(w http.ResponseWriter, r *http.Request) {
		var conn *websocket.Conn
		var err error
		if conn, err = websocket.AcceptHTTP(w, r); err != nil {
			// this might not even work
			w.WriteHeader(500)
			w.Write([]byte("failed to accept websocket"))
			slog.Error("failed to accept websocket", "err", err.Error())
			return
		}
		liveRequestWebsockets = append(liveRequestWebsockets, conn)
	})

	http.HandleFunc("OPTIONS /", func(w http.ResponseWriter, _ *http.Request) {
		setCORSHeaders(w)
		w.WriteHeader(http.StatusNoContent)
	})

	go func() {
		for msg := range liveRequestMessages {
			newLiveRequestWebsockets := liveRequestWebsockets
			for i, ws := range liveRequestWebsockets {
				err := ws.Write(&websocket.Message{
					Type: websocket.MessageText,
					Data: msg,
				})
				if err != nil {
					ws.Close() // close the conn, but it might be already closed
					// do NOT use slices.Delete here. it will cause a nil panic because of the clearing in slices.Delete
					newLiveRequestWebsockets = append(liveRequestWebsockets[:i], liveRequestWebsockets[i+1:]...)
				}
			}
			liveRequestWebsockets = newLiveRequestWebsockets // update the slice so it doesn't grow indefinitely or mess with the range loop
		}
	}()

	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		panic(err)
	}
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Request-Method", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Max-Age", "300")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

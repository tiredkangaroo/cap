package main

import (
	"encoding/json"
	"io"
	"log/slog"

	nethttp "net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
	"github.com/tiredkangaroo/websocket"
)

// we will continue to use net/http for the control server

func startControlServer(m *Manager, ph *ProxyHandler) {
	mux := nethttp.NewServeMux()

	if config.DefaultConfig.Debug {
		slog.Info("debug mode enabled")
		mux.Handle("GET /debug/pprof/", nethttp.DefaultServeMux)
		mux.Handle("GET /debug/pprof/cmdline", nethttp.DefaultServeMux)
		mux.Handle("GET /debug/pprof/profile", nethttp.DefaultServeMux)
		mux.Handle("GET /debug/pprof/symbol", nethttp.DefaultServeMux)
		mux.Handle("GET /debug/pprof/trace", nethttp.DefaultServeMux)
	} else {
		slog.Info("debug mode disabled, use DEBUG=true to enable it")
	}

	// Start the control server
	mux.HandleFunc("GET /config", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		setCORSHeaders(w)

		data, err := json.Marshal(config.DefaultConfig)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("failed to marshal config"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		w.Write(data)
	})

	mux.HandleFunc("POST /config", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		setCORSHeaders(w)

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
		w.WriteHeader(nethttp.StatusOK)
		w.Write([]byte("config updated"))
	})

	mux.HandleFunc("GET /requestsWS", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		var conn *websocket.Conn
		var err error
		if conn, err = m.AcceptWS(w, r); err != nil {
			// this might not even work
			w.WriteHeader(500)
			w.Write([]byte("failed to accept websocket"))
			slog.Error("failed to accept websocket", "err", err.Error())
			return
		}
		// NOTE: if there's a write error, and its deleted, what happens? grtine leak?
		go func() {
			for {
				msg, err := conn.Read()
				if err != nil {
					slog.Error("failed to read from websocket", "err", err.Error())
					return
				}
				m.HandleMessage(msg)
			}
		}()
	})

	// nethttp.HandleFunc("GET /response", func (w nethttp.ResponseWriter, r *nethttp.Request) {

	// })
	mux.HandleFunc("GET /request/{id}", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		setCORSHeaders(w)
		id := r.PathValue("id")
		if id == "" {
			w.WriteHeader(nethttp.StatusBadRequest)
			w.Write([]byte("missing id parameter"))
			return
		}
		req, err := m.db.GetRequestByID(id)
		if err != nil {
			w.WriteHeader(nethttp.StatusNotFound)
			w.Write([]byte("request not found"))
			slog.Error("failed to get request by ID", "id", id, "err", err.Error())
			return
		}
		data, err := json.Marshal(req)
		if err != nil {
			w.WriteHeader(nethttp.StatusInternalServerError)
			w.Write([]byte("failed to marshal request data"))
			slog.Error("failed to marshal request data", "id", id, "err", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		w.Write(data)
	})

	mux.HandleFunc("OPTIONS /", func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		setCORSHeaders(w)
		w.WriteHeader(nethttp.StatusNoContent)
	})

	mux.HandleFunc("GET /filterCounts", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		setCORSHeaders(w)

		counts, err := m.db.GetFilterCounts()
		if err != nil {
			w.WriteHeader(nethttp.StatusInternalServerError)
			w.Write([]byte("failed to get filter counts"))
			slog.Error("failed to get filter counts", "err", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		w.Write(marshal(counts))
	})

	mux.HandleFunc("GET /requestsMatchingFilter", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		setCORSHeaders(w)

		query := r.URL.Query()
		offset := query.Get("offset")
		limit := query.Get("limit")

		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			w.Write([]byte("invalid offset parameter"))
			return
		}
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			w.Write([]byte("invalid limit parameter"))
			return
		}
		if limitInt <= 0 || offsetInt < 0 {
			w.WriteHeader(nethttp.StatusBadRequest)
			w.Write([]byte("offset must be > 0 and limit must be > 0"))
			return
		}

		filter := Filter{
			ClientApplication: query.Get("clientApplication"),
			Host:              query.Get("host"),
			ClientIP:          query.Get("clientIP"),
		}

		paginatedRequests, totalRequests, err := m.db.GetRequestsMatchingFilter(filter, offsetInt, limitInt)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			w.Write([]byte("failed to get requests"))
			slog.Error("failed to get requests matching filter", "err", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		w.Write(marshal(map[string]any{
			"requests": paginatedRequests,
			"total":    totalRequests,
		}))
	})

	err := nethttp.ListenAndServe(":8001", mux)
	if err != nil {
		panic(err)
	}
}

func setCORSHeaders(w nethttp.ResponseWriter) {
	if !config.DefaultConfig.Debug {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Request-Method", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Max-Age", "300")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

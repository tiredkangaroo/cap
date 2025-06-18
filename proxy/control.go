package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
	"github.com/tiredkangaroo/websocket"
)

func startControlServer(m *Manager, ph *ProxyHandler) {
	// Start the control server
	http.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)

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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("config updated"))
	})

	http.HandleFunc("GET /requestsWS", func(w http.ResponseWriter, r *http.Request) {
		var conn *websocket.Conn
		var err error
		if conn, err = m.AcceptWS(w, r); err != nil {
			// this might not even work
			w.WriteHeader(500)
			w.Write([]byte("failed to accept websocket"))
			slog.Error("failed to accept websocket", "err", err.Error())
			return
		}
		//NOTE: if there's a write error, and its deleted, what happens? grtine leak?
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

	// http.HandleFunc("GET /response", func (w http.ResponseWriter, r *http.Request) {

	// })

	http.HandleFunc("GET /repeat/{id}", func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)
		id := r.PathValue("id")
		// prevent resource exhaustion by limiting the size of the request body (but it likely won't be an issue because
		// this control server isn't outward facing)
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing id parameter"))
			return
		}
		req, err := m.db.GetRequestByID(id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("request not found"))
			return
		}
		// we will add init steps here to create a new request
		// and then we'll use our designated handlers (get kind to choose which handler to use)

		newReq := new(Request)
		*newReq = *req
		newReq.req = req.req.Clone(r.Context())
		newReq.req.Host = req.Host
		if err := newReq.Init(w, newReq.req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to initialize request"))
			slog.Error("failed to initialize request", "id", id, "err", err.Error())
			return
		}
		defer newReq.conn.Close()
		newReq.Kind = req.Kind
		// newReq.Secure = req.Kind == RequestKindHTTPS || req.Kind == RequestKindHTTPSMITM
		ph.serveAfterInit(newReq)
	})

	http.HandleFunc("GET /request/{id}", func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)
		id := r.PathValue("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing id parameter"))
			return
		}
		req, err := m.db.GetRequestByID(id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("request not found"))
			slog.Error("failed to get request by ID", "id", id, "err", err.Error())
			return
		}
		data, err := json.Marshal(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to marshal request data"))
			slog.Error("failed to marshal request data", "id", id, "err", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	http.HandleFunc("OPTIONS /", func(w http.ResponseWriter, _ *http.Request) {
		setCORSHeaders(w)
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("GET /filterCounts", func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)

		counts, err := m.db.GetFilterCounts()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to get filter counts"))
			slog.Error("failed to get filter counts", "err", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(marshal(counts))
	})

	http.HandleFunc("GET /requestsMatchingFilter", func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)

		query := r.URL.Query()
		offset := query.Get("offset")
		limit := query.Get("limit")

		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid offset parameter"))
			return
		}
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid limit parameter"))
			return
		}
		if limitInt <= 0 || offsetInt < 0 {
			w.WriteHeader(http.StatusBadRequest)
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
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to get requests"))
			slog.Error("failed to get requests matching filter", "err", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(marshal(map[string]any{
			"requests": paginatedRequests,
			"total":    totalRequests,
		}))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)
		cookie, err := r.Cookie("CAP-PROXIED-HOST")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("you're probably here by mistake"))
			return
		}
		if cookie.Value == "" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("you're probably here by mistake"))
			return
		}
		newReq := new(Request)
		newReq.req.Host = cookie.Value
		newReq.req.URL = r.URL
		if err := newReq.Init(w, newReq.req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to initialize request"))
			return
		}
		ph.ServeHTTP(w, r)
	})

	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		panic(err)
	}
}

func setCORSHeaders(w http.ResponseWriter) {
	if !config.DefaultConfig.Debug {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Request-Method", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Max-Age", "300")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

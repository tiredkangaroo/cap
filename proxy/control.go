package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"

	nethttp "net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/tiredkangaroo/cap/proxy/config"
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

	mux.HandleFunc("GET /reqbody/{id}", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		id := r.PathValue("id")
		if id == "" {
			w.WriteHeader(nethttp.StatusBadRequest)
			w.Write([]byte("missing id parameter"))
			return
		}

		hijacker := w.(nethttp.Hijacker)
		conn, _, err := hijacker.Hijack()
		if err != nil {
			w.WriteHeader(nethttp.StatusInternalServerError)
			w.Write([]byte("failed to hijack connection"))
			slog.Error("failed to hijack connection", "id", id, "err", err.Error())
			return
		}
		defer conn.Close()
		conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		conn.Write([]byte("Content-Type: text/plain\r\n"))
		writeRawCORSHeaders(conn)

		if waiter, ok := m.approvalWaiters[id]; ok {
			if waiter.req != nil && waiter.req.Body != nil { // jic to avoid npd panics but the second clause shoud always be true if the first one is
				conn.Write(fmt.Appendf([]byte{}, "Content-Length: %d\r\n\r\n", waiter.req.Body.ContentLength()))
				if _, err := waiter.req.Body.WriteTo(conn); err != nil {
					slog.Error("failed to write request body", "id", id, "err", err.Error())
					conn.Write([]byte("failed to write request body"))
				} else {
					slog.Info("wrote request body", "id", id)
				}
				return
			} else {
				conn.Write([]byte("req body unavailable"))
				return
			}
		}

		err = m.db.WriteRequestBody(id, NewNoOpCloser(conn))
		if err != nil {
			slog.Error("failed to write request body", "id", id, "err", err.Error())
			return
		}
	})

	mux.HandleFunc("GET /respbody/{id}", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		id := r.PathValue("id")
		if id == "" {
			w.WriteHeader(nethttp.StatusBadRequest)
			w.Write([]byte("missing id parameter"))
			return
		}

		hijacker := w.(nethttp.Hijacker)
		conn, _, err := hijacker.Hijack()
		if err != nil {
			w.WriteHeader(nethttp.StatusInternalServerError)
			w.Write([]byte("failed to hijack connection"))
			slog.Error("failed to hijack connection", "id", id, "err", err.Error())
			return
		}
		defer conn.Close()
		conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		conn.Write([]byte("Content-Type: text/plain\r\n"))
		writeRawCORSHeaders(conn)
		err = m.db.WriteResponseBody(id, NewNoOpCloser(conn))
		if err != nil {
			w.WriteHeader(nethttp.StatusInternalServerError)
			w.Write([]byte("failed to write request body"))
			slog.Error("failed to write request body", "id", id, "err", err.Error())
			return
		}
	})

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

	mux.HandleFunc("GET /filter", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		setCORSHeaders(w)

		filter := Filter{
			FilterField{
				Name:        "clientApplication",
				Type:        FilterTypeString,
				VerboseName: "Client Application",
			},
			FilterField{
				Name:        "host",
				Type:        FilterTypeString,
				VerboseName: "Host",
			},
			FilterField{
				Name:        "clientIP",
				Type:        FilterTypeString,
				VerboseName: "Client IP",
			},
			FilterField{
				Name:        "starred",
				Type:        FilterTypeBool,
				VerboseName: "Starred Only",
			},
		}

		err := m.db.GetFilterUniqueValues(filter)
		if err != nil {
			w.WriteHeader(nethttp.StatusInternalServerError)
			w.Write([]byte("failed to get filter counts"))
			slog.Error("failed to get filter counts", "err", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		w.Write(marshal(filter))
	})

	mux.HandleFunc("POST /setRequestStarred/{id}", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		setCORSHeaders(w)

		id := r.PathValue("id")

		starred, err := strconv.ParseBool(r.URL.Query().Get("starred"))
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			w.Write([]byte("invalid starred parameter"))
			return
		}

		err = m.db.SetRequestStarred(id, starred)
		if err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			w.Write([]byte("failed to set request starred status"))
			slog.Error("failed to set request starred status", "id", id, "err", err.Error())
			return
		}

		w.WriteHeader(nethttp.StatusOK)
		w.Write([]byte("request starred status updated"))
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
			FilterField{
				Name:          "clientApplication",
				Type:          FilterTypeString,
				UniqueValues:  nil,
				SelectedValue: query.Get("clientApplication"),
			},
			FilterField{
				Name:          "host",
				Type:          FilterTypeString,
				UniqueValues:  nil,
				SelectedValue: query.Get("host"),
			},
			FilterField{
				Name:          "clientIP",
				Type:          FilterTypeString,
				UniqueValues:  nil,
				SelectedValue: query.Get("clientIP"),
			},
		}
		if query.Get("starred") != "" {
			starred, err := strconv.ParseBool(query.Get("starred"))
			if err != nil {
				w.WriteHeader(nethttp.StatusBadRequest)
				w.Write([]byte("invalid starred parameter"))
				return
			}
			filter = append(filter, FilterField{
				Name:          "starred",
				Type:          FilterTypeBool,
				UniqueValues:  nil,
				SelectedValue: starred,
			})
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

func writeRawCORSHeaders(c net.Conn) {
	if !config.DefaultConfig.Debug {
		return
	}
	c.Write([]byte("Access-Control-Allow-Origin: *\r\n"))
	c.Write([]byte("Access-Control-Request-Method: POST, GET, OPTIONS\r\n"))
	c.Write([]byte("Access-Control-Max-Age: 300\r\n"))
	c.Write([]byte("Access-Control-Allow-Headers: Content-Type\r\n"))
}

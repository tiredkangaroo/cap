package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
	"github.com/tiredkangaroo/websocket"
)

func startControlServer(controlMessages *ControlChannel) {
	controlMessagesWebsockets := []*websocket.Conn{}
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
		controlMessagesWebsockets = append(controlMessagesWebsockets, conn)
		//NOTE: if there's a write error, and its deleted, what happens? grtine leak?
		go func() {
			for {
				msg, err := conn.Read()
				if err != nil {
					slog.Error("failed to read from websocket", "err", err.Error())
					return
				}
				handleClientMessage(msg, controlMessages)
			}
		}()
	})

	http.HandleFunc("OPTIONS /", func(w http.ResponseWriter, _ *http.Request) {
		setCORSHeaders(w)
		w.WriteHeader(http.StatusNoContent)
	})

	go func() {
		for msg := range controlMessages.u {
			newLiveRequestWebsockets := controlMessagesWebsockets
			for _, ws := range controlMessagesWebsockets {
				err := ws.Write(&websocket.Message{
					Type: websocket.MessageText,
					Data: msg,
				})
				if err != nil {
					slog.Error("failed to write to websocket", "ws", ws, "err", err.Error())
					// FIXME: a working connection might "fail to write" but its not even closed yet and can be perfectly used?! also the
					// code works even if i delete the websocket from the slice on this error condition, and the ws is still perfectly usable
					// somehow for writing again and syncs to the clietn
					//
					// i dont understand :( :(  :(
					// ws.Close() // close the conn, but it might be already closed
					// do NOT use slices.Delete here. it will cause a nil panic because of the clearing in slices.Delete
					// newLiveRequestWebsockets = append(controlMessagesWebsockets[:i], controlMessagesWebsockets[i+1:]...)
				}
			}
			controlMessagesWebsockets = newLiveRequestWebsockets // update the slice so it doesn't grow indefinitely or mess with the range loop
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

func handleClientMessage(msg *websocket.Message, controlMessages *ControlChannel) {
	if msg.Type != websocket.MessageText {
		slog.Error("invalid message type", "type", msg.Type)
		return
	}
	sData := strings.Split(string(msg.Data), " ")
	if len(sData) < 2 {
		slog.Error("invalid control message", "data", string(msg.Data))
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(strings.Join(sData[1:], " ")), &data); err != nil {
		slog.Error("failed to unmarshal control message", "data", string(msg.Data), "err", err.Error())
		return
	}
	switch sData[0] {
	case "WAIT-APPROVAL-APPROVE":
		if len(sData) < 2 {
			slog.Error("invalid control message", "data", string(msg.Data))
			return
		}
		id, ok := data["id"].(string)
		if !ok {
			slog.Error("invalid control message", "data", string(msg.Data), "err", "id is not a string")
			return
		}
		controlMessages.mxWaitingApprovalResponse.Lock()
		if fn, ok := controlMessages.waitingApprovalResponse[id]; ok {
			slog.Info("approval received for request", "id", id)
			fn(true)
		}
		controlMessages.mxWaitingApprovalResponse.Unlock()
	case "WAIT-APPROVAL-CANCELED":
		if len(sData) < 2 {
			slog.Error("invalid control message", "data", string(msg.Data))
			return
		}
		id, ok := data["id"].(string)
		if !ok {
			slog.Error("invalid control message", "data", string(msg.Data), "err", "id is not a string")
			return
		}
		controlMessages.mxWaitingApprovalResponse.Lock()
		if fn, ok := controlMessages.waitingApprovalResponse[id]; ok {
			slog.Info("cancel received for request", "id", id)
			fn(false)
		}
		controlMessages.mxWaitingApprovalResponse.Unlock()

	}
}

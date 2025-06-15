package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/tiredkangaroo/websocket"
)

// message defined:
//
// (proxy -> client) NEW <id, datetime, host, secure, clientIP, clientAuthorization>
// (proxy -> client) REQUEST <id, method, path, query, headers, body>
// (proxy -> client) APPROVAL-WAIT <id>
// (client -> proxy) APPROVAL-APPROVE <id>
// (client -> proxy) APPROVAL-CANCEL <id>
// (proxy -> client) APPROVAL-RECIEVED <id>
// (proxy -> client) APPROVAL-CANCELED <id>
// (proxy -> client) RESPONSE <id, statusCode, headers, body>
// (proxy -> client) DONE <id>
// (proxy -> client) ERROR <id, err>

type Manager struct {
	db      *Database
	wsConns []*websocket.Conn

	approvalWaiters     map[string]*Request
	approvalWaitersRWMu sync.RWMutex
}

type IDMessage struct {
	ID string `json:"id"`
}

func (c *Manager) AcceptWS(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := websocket.AcceptHTTP(w, r)
	if err != nil {
		return nil, err
	}
	c.wsConns = append(c.wsConns, conn)
	return conn, nil
}

func (c *Manager) SendNew(req *Request) {
	var secureState string
	switch req.Kind {
	case RequestKindHTTP:
		secureState = "HTTP (Insecure)"
	case RequestKindHTTPS:
		secureState = "HTTPS (Secure)"
	case RequestKindHTTPSMITM:
		secureState = "HTTPS (with MITM)"
	}
	c.writeJSON("NEW", map[string]any{
		"id":                  req.ID,
		"datetime":            req.Datetime,
		"host":                req.Host,
		"secureState":         secureState,
		"clientIP":            req.ClientIP,
		"clientAuthorization": req.ClientAuthorization,
		"clientProcessID":     req.ClientProcessID,
		"clientApplication":   req.ClientProcessName,
	})
}

func (c *Manager) SendTunnel(req *Request) {
	c.writeJSON("TUNNEL", IDMessage{
		ID: req.ID,
	})
}

func (c *Manager) SendRequest(req *Request) {
	// NOTE: HTTP requests will have a 0 bytesTransferred value because the request is read before the connection is hijacked and
	// made custom.
	c.writeJSON("REQUEST", map[string]any{
		"id":               req.ID,
		"method":           req.req.Method,
		"path":             req.req.URL.Path,
		"query":            req.req.URL.Query(),
		"headers":          req.req.Header,
		"body":             string(req.body()),
		"bytesTransferred": req.BytesTransferred(),
	})
}

func (c *Manager) SendResponse(req *Request) {
	c.writeJSON("RESPONSE", map[string]any{
		"id":         req.ID,
		"statusCode": req.resp.StatusCode,
		"headers":    req.resp.Header,
		"body":       string(req.respbody()),
	})
}

// NOTE: add timing_total to timing.export
func (c *Manager) SendDone(req *Request) {
	if err := c.db.SaveRequest(req, nil); err != nil {
		slog.Error("saving done request to database", "err", err.Error(), "request_id", req.ID)
	}
	c.writeJSON("DONE", map[string]any{
		"id":               req.ID,
		"bytesTransferred": req.BytesTransferred(),
		"timing":           req.timing.Export(),
		"timing_total":     req.timing.Total(),
	})
}

func (c *Manager) SendError(req *Request, err error) {
	if e := c.db.SaveRequest(req, err); e != nil {
		slog.Error("saving erroneous request to database", "err", e.Error(), "request_id", req.ID)
	}
	c.writeJSON("ERROR", map[string]any{
		"id":               req.ID,
		"error":            err.Error(),
		"bytesTransferred": req.BytesTransferred(), //NOTE: field not handled yet
		"timing":           req.timing.Export(),
		"timing_total":     req.timing.Total(),
	})
}

// RecieveApproval waits for an approval request from the client. It blocks until the client approves or cancels the request.
// If the client approves, it returns true, otherwise it returns false.
func (c *Manager) RecieveApproval(req *Request) (approved bool) {
	ctx, cancel := context.WithCancel(context.Background())

	c.approvalWaitersRWMu.Lock()
	c.approvalWaiters[req.ID] = req
	c.approvalWaitersRWMu.Unlock()

	req.approveResponseFunc = func(approval bool) {
		defer cancel()
		approved = approval
		if approved {
			c.writeJSON("APPROVAL-RECIEVED", IDMessage{
				ID: req.ID,
			})
		} else {
			c.writeJSON("APPROVAL-CANCELED", IDMessage{
				ID: req.ID,
			})
		}
	}

	c.writeJSON("APPROVAL-WAIT", IDMessage{
		ID: req.ID,
	})

	<-ctx.Done()
	return approved
}

func (c *Manager) HandleMessage(msg *websocket.Message) {
	if msg.Type != websocket.MessageText {
		return // we're not interested in other message types
	}

	parts := bytes.SplitN(msg.Data, []byte(" "), 2)
	if len(parts) < 2 {
		return // invalid message format
	}
	action := string(parts[0])
	data := parts[1]

	switch action {
	case "APPROVAL-APPROVE":
		c.handleApprovalApprove(data)
	case "APPROVAL-CANCEL":
		c.handleApprovalCancel(data)
	case "UPDATE-REQUEST":
		c.handleUpdateRequest(data)
	}
}

func (c *Manager) handleApprovalApprove(data []byte) {
	req, ok := c.getApprovalWaitingRequestFromIDMessage(data)
	if !ok {
		// handle error: request not found
		return
	}
	req.approveResponseFunc(true)
}

func (c *Manager) handleApprovalCancel(data []byte) {
	req, ok := c.getApprovalWaitingRequestFromIDMessage(data)
	if !ok {
		// handle error: request not found
		return
	}
	req.approveResponseFunc(false)
}

func (c *Manager) handleUpdateRequest(data []byte) {
	fmt.Println("166", string(data))
	type updatedMessageType struct {
		IDMessage
		Request struct {
			Body    string      `json:"body"`
			Headers http.Header `json:"headers"`
			Host    string      `json:"host"`
			Method  string      `json:"method"`
			Path    string      `json:"path"`
			Query   url.Values  `json:"query"`
		} `json:"request"`
	}
	updatedMessage, err := expectJSON[updatedMessageType](data)
	if err != nil {
		// handle error: invalid message format
		return
	}

	c.approvalWaitersRWMu.RLock()
	defer c.approvalWaitersRWMu.RUnlock()
	req, ok := c.approvalWaiters[updatedMessage.ID]
	if !ok {
		// handle error: request not found
		return
	}
	if req.req.Body != nil {
		req.req.Body.Close() // close the old body if it exists
	}

	req.req.Body = io.NopCloser(strings.NewReader(updatedMessage.Request.Body))
	req.req.Header = updatedMessage.Request.Headers
	req.req.Host = updatedMessage.Request.Host
	req.req.Method = updatedMessage.Request.Method
	req.req.URL.Host = updatedMessage.Request.Host
	req.req.URL.Path = updatedMessage.Request.Path
	req.req.URL.RawQuery = updatedMessage.Request.Query.Encode()

}

// getApprovalWaitingRequestFromIDMessage retrieves the request associated with the given ID message with the map
// for approval waiters. It returns the request and a boolean indicating success. It will delete the request from the map
// if it is found.
func (c *Manager) getApprovalWaitingRequestFromIDMessage(data []byte) (*Request, bool) {
	idMessage, err := expectJSON[IDMessage](data)
	if err != nil {
		// handle error: invalid message format
		return nil, false
	}

	c.approvalWaitersRWMu.Lock()
	defer c.approvalWaitersRWMu.Unlock()
	req, ok := c.approvalWaiters[idMessage.ID]

	if !ok {
		// handle error: request not found
		return nil, false
	}

	delete(c.approvalWaiters, idMessage.ID)
	return req, true
}

func (c *Manager) writeJSON(action string, data any) {
	d, _ := json.Marshal(data)
	for _, conn := range c.wsConns {
		conn.Write(&websocket.Message{
			Type: websocket.MessageText,
			Data: append([]byte(action+" "), d...),
		})
	}
}

func expectJSON[T any](data []byte) (T, error) {
	var v T
	err := json.Unmarshal(data, &v)
	if err != nil {
		return v, err
	}
	return v, nil
}

package main

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"sync"
)

type ControlChannel struct {
	u                         chan []byte
	mxWaitingApprovalResponse sync.Mutex
	waitingApprovalResponse   map[string]*Request // maps req id to request
}

// toURL converts a string to a URL. If the string does not start with "http://" or
// "https://", it will prepend "http://" or "https://" based on the https parameter.
// It returns an error if the string is not a valid URL after conversion.
//
// This function is used to ensure that the URL is in a valid format before
// performing any net/http Do operations.
func toURL(s string, https bool) (*url.URL, error) {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		if https {
			s = "https://" + s
		} else {
			s = "http://" + s
		}
	}
	return url.Parse(s)
}

func (c *ControlChannel) sendControlNew(req *Request) {
	data, _ := json.Marshal(map[string]any{
		"id":                  req.id,
		"datetime":            req.datetime.Format("2006-01-02 15:04:05"),
		"secure":              req.secure,
		"clientIP":            req.clientIP,
		"clientAuthorization": req.clientAuthorization,
		"host":                req.host,
	})
	c.u <- append([]byte("NEW "), data...)
}

func (c *ControlChannel) sendControlHTTPRequest(req *Request) {
	data, _ := json.Marshal(map[string]any{
		"id":      req.id,
		"method":  req.req.Method,
		"path":    req.req.URL.Path,
		"query":   req.req.URL.Query(),
		"headers": req.req.Header,
		"body":    string(req.body()),
	})
	c.u <- append([]byte("HTTP "), data...)
}

func (c *ControlChannel) sendControlHTTPSMITMRequest(req *Request) {
	data, _ := json.Marshal(map[string]any{
		"id":      req.id,
		"method":  req.req.Method,
		"path":    req.req.URL.Path,
		"query":   req.req.URL.Query(),
		"headers": req.req.Header,
		"body":    string(req.body()),
	})
	c.u <- append([]byte("HTTPS-MITM-REQUEST "), data...)
}

func (c *ControlChannel) sendControlHTTPSTunnelRequest(req *Request) {
	data, _ := json.Marshal(map[string]any{
		"id": req.id,
	})
	c.u <- append([]byte("HTTPS-TUNNEL-REQUEST "), data...)
}

func (c *ControlChannel) sendControlHTTPResponse(req *Request) {
	data, _ := json.Marshal(map[string]any{
		"id":         req.id,
		"statusCode": req.resp.StatusCode,
		"headers":    req.resp.Header,
		"body":       string(req.respbody()),
	})
	c.u <- append([]byte("HTTP-RESPONSE "), data...)
}

func (c *ControlChannel) sendControlHTTPSMITMResponse(req *Request) {
	data, _ := json.Marshal(map[string]any{
		"id":         req.id,
		"statusCode": req.resp.StatusCode,
		"headers":    req.resp.Header,
		"body":       string(req.respbody()),
	})
	c.u <- append([]byte("HTTPS-MITM-RESPONSE "), data...)
}

func (c *ControlChannel) sendControlHTTPSTunnelResponse(req *Request) {
	data, _ := json.Marshal(map[string]any{
		"id": req.id,
	})
	c.u <- append([]byte("HTTPS-TUNNEL-RESPONSE "), data...)
}

func (c *ControlChannel) sendControlDone(req *Request) {
	data, _ := json.Marshal(map[string]any{
		"id": req.id,
	})
	c.u <- append([]byte("DONE "), data...)
}

func (c *ControlChannel) sendControlError(req *Request, err error) {
	data, _ := json.Marshal(map[string]any{
		"id":    req.id,
		"error": err.Error(),
	})
	c.u <- append([]byte("ERROR "), data...)
}

func (c *ControlChannel) waitApproval(req *Request) bool {
	data, _ := json.Marshal(map[string]any{
		"id": req.id,
	})
	// NOTE: default approval timeout here

	var cancelApproval bool
	ctx, cancel := context.WithCancel(context.Background())

	req.approveResponseFunc = func(approved bool) {
		cancelApproval = approved
		cancel() // cancel the context to stop waiting
	}

	c.mxWaitingApprovalResponse.Lock()
	c.waitingApprovalResponse[req.id] = req
	c.mxWaitingApprovalResponse.Unlock()
	c.u <- append([]byte("WAIT-APPROVAL "), data...)

	<-ctx.Done()
	switch ctx.Err() {
	case context.Canceled:
		c.mxWaitingApprovalResponse.Lock()
		delete(c.waitingApprovalResponse, req.id)
		c.mxWaitingApprovalResponse.Unlock()

		if cancelApproval {
			// approval was given
			c.u <- append([]byte("WAIT-APPROVAL-APPROVED "), data...)
			return true
		} else {
			c.u <- append([]byte("CANCELED "), data...)
			return false
		}
	case context.DeadlineExceeded: // unreachable for now, but infra is in place to handle this
		// approval was not given in time
		c.u <- append([]byte("WAIT-APPROVAL-TIMEOUT "), data...)
		return false
	default:
		// unreachable but we walkin
		return false
	}
}

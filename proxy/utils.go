package main

import (
	"encoding/json"
	"net/url"
	"strings"
)

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

func sendControlNew(req *Request, c chan []byte) {
	data, _ := json.Marshal(map[string]any{
		"id":                  req.id,
		"secure":              req.secure,
		"clientIP":            req.clientIP,
		"clientAuthorization": req.clientAuthorization,
		"host":                req.host,
	})
	c <- append([]byte("NEW "), data...)
}

func sendControlHTTPRequest(req *Request, c chan []byte) {
	data, _ := json.Marshal(map[string]any{
		"id":      req.id,
		"method":  req.req.Method,
		"path":    req.req.URL.Path,
		"headers": req.req.Header,
		"body":    string(req.body()),
	})
	c <- append([]byte("HTTP "), data...)
}

func sendControlHTTPSMITMRequest(req *Request, c chan []byte) {
	data, _ := json.Marshal(map[string]any{
		"id":      req.id,
		"method":  req.req.Method,
		"path":    req.req.URL.Path,
		"headers": req.req.Header,
		"body":    string(req.body()),
	})
	c <- append([]byte("HTTPS-MITM-REQUEST "), data...)
}

func sendControlHTTPSTunnelRequest(req *Request, c chan []byte) {
	// NOTE: this will later includes bytes transferred etc. but also not just for no MITM both an provide that info
	data, _ := json.Marshal(map[string]any{
		"id": req.id,
	})
	c <- append([]byte("HTTPS-TUNNEL-REQUEST "), data...)
}

func sendControlHTTPResponse(req *Request, c chan []byte) {
	data, _ := json.Marshal(map[string]any{
		"id":         req.id,
		"statusCode": req.resp.StatusCode,
		"headers":    req.resp.Header,
		"body":       string(req.respbody()),
	})
	c <- append([]byte("HTTP-RESPONSE "), data...)
}

func sendControlHTTPSMITMResponse(req *Request, c chan []byte) {
	data, _ := json.Marshal(map[string]any{
		"id":         req.id,
		"statusCode": req.resp.StatusCode,
		"headers":    req.resp.Header,
		"body":       string(req.respbody()),
	})
	c <- append([]byte("HTTPS-MITM-RESPONSE "), data...)
}

func sendControlHTTPSTunnelResponse(req *Request, c chan []byte) {
	data, _ := json.Marshal(map[string]any{
		"id": req.id,
	})
	c <- append([]byte("HTTPS-TUNNEL-RESPONSE "), data...)
}

func sendControlDone(req *Request, c chan []byte) {
	data, _ := json.Marshal(map[string]any{
		"id": req.id,
	})
	c <- append([]byte("DONE "), data...)
}

func sendControlError(req *Request, err error, c chan []byte) {
	data, _ := json.Marshal(map[string]any{
		"id":    req.id,
		"error": err.Error(),
	})
	c <- append([]byte("ERROR "), data...)
}

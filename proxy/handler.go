package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	certificate "github.com/tiredkangaroo/bigproxy/proxy/certificates"
	"github.com/tiredkangaroo/bigproxy/proxy/config"
	"github.com/tiredkangaroo/bigproxy/proxy/http"
	"github.com/tiredkangaroo/bigproxy/proxy/timing"
)

var (
	// ErrResponseWriterNoHijack is returned when the ResponseWriter does not implement the Hijacker interface.
	ErrResponseWriterNoHijack = errors.New("response writer cannot be hijacked: does not implement Hijacker interface")
)
var (
	// ResponseHijackingError is the response sent to the client when the proxy cannot hijack the connection.
	ResponseHijackingError []byte = []byte("hijacking error")

	// ResponseRawSuccess is the response sent to the client when the HTTPS proxy successfully hijacks the connection.
	ResponseRawSuccess []byte = []byte("HTTP/1.1 200 OK\r\n" +
		"Connection: close\r\n" +
		"Content-Length: 0\r\n" +
		"\r\n")
)

// look over HTTP/2 handling and streaming responses

// handleHTTP handles a HTTP request to the proxy.
//
// HTTP requests do not create secure tunnels, therefore the original request to the
// proxy is the one meant for the host. The only prep we need to do is strip proxy
// headers. The proxy will then perform the request to the host and send the response
// back to the client.
func (r *Request) handleHTTP(m *Manager, req *http.Request, c *certificate.Certificates) error {
	// HTTP requests send the full request to the proxy, so this is the request we want to perform
	r.req = req

	// send request to live websocket connections
	m.SendRequest(r)

	// perform the request
	resp, err := r.Perform(m, c)
	if err != nil {
		return fmt.Errorf("perform: %w", err)
	}

	// send response to live websocket connections
	m.SendResponse(r)

	// save the request body to the database
	r.timing.Start(timing.TimeSaveRequestBody)
	if err := m.db.SaveBody(r.reqBodyID, r.req.Body); err != nil {
		slog.Error("save request body: %w", "err", err)
	} else {
		slog.Debug("saved request body", "id", r.reqBodyID)
	}
	r.timing.Stop()

	// save the response body to the database
	r.timing.Start(timing.TimeSaveResponseBody)
	if err := m.db.SaveBody(r.respBodyID, r.resp.Body); err != nil {
		slog.Error("save response body: %w", "err", err)
	} else {
		slog.Debug("saved response body", "id", r.reqBodyID)
	}
	r.timing.Stop()

	// write the response to the connection
	r.timing.Start(timing.TimeWriteResponse)
	err = resp.Write(r.conn) // write the response to the connection
	r.timing.Stop()
	if err != nil {
		return fmt.Errorf("connection write: %w", err)
	}

	return nil
}

// handleHTTPS handles a HTTPS request. This proxy is a MITM proxy, so it will
// use a self-signed certificate for the requested host and send it to the client.
//
// Steps:
// (1) It writes a success response to the client. The client believes it is no longer speaking to the proxy now.
// (2) A TLS handshake (using the self-signed) occurs between the client and the proxy. The client believes it is speaking to the host.
// (3) The proxy reads the actual request from the client that it meant to send the host.
// (4) The proxy performs the request to the real host.
// (5) The proxy sends the response back to the client.
func (r *Request) handleHTTPS(m *Manager, c *certificate.Certificates) error {
	// write a success response to the client (this is meant to be the last thing before the secure tunnel is expected)
	r.timing.Start(timing.TimeSendProxyResponse)
	_, err := r.conn.Write(ResponseRawSuccess)
	r.timing.Stop()
	if err != nil {
		return fmt.Errorf("connection write: %w", err)
	}

	if !config.DefaultConfig.MITM {
		return r.handleNoMITM(m)
	} else if c == nil {
		return fmt.Errorf("mitm is enabled, but certificate service is unavailable")
	}

	// after the success response, a handshake will occur and the user will
	// send the ACTUAL request.
	// we need to perform a TLS handshake with the client using the self-signed certificate for the requested host.
	// this will allow us to read the request from the client as if we were the host.

	// NOTE: consider IPV6 square bracket and how that affects the hostname
	r.timing.Start(timing.TimeCertGenTLSHandshake)
	tlsconn, err := c.TLSConn(r.conn, getHostname(r.Host))
	if err != nil {
		return fmt.Errorf("tls conn: %w", err)
	}
	if err := tlsconn.Handshake(); err != nil {
		return fmt.Errorf("tls handshake: %w", err)
	}
	r.timing.Stop()

	// read the request from the TLS connection (this is the ACTUAL request meant for the host, which we will perform)
	r.timing.Start(timing.TimeReadRequest)
	req, err := http.ReadRequest(tlsconn)
	r.timing.Stop()
	if err != nil {
		return fmt.Errorf("read mitm request: %w", err)
	}
	r.req = req

	// send the request to live websocket connections
	m.SendRequest(r)

	resp, err := r.Perform(m, c)
	if err != nil {
		return fmt.Errorf("perform: %w", err)
	}

	// send the response to live websocket connections
	m.SendResponse(r)

	// save the request body to the database
	r.timing.Start(timing.TimeSaveRequestBody)
	if err := m.db.SaveBody(r.reqBodyID, r.req.Body); err != nil {
		slog.Error("save request body: %w", "err", err)
	} else {
		slog.Debug("saved request body", "id", r.reqBodyID)
	}
	r.timing.Stop()

	// save the response body to the database
	r.timing.Start(timing.TimeSaveResponseBody)
	if err := m.db.SaveBody(r.respBodyID, resp.Body); err != nil {
		slog.Error("save response body: %w", "err", err)
	} else {
		slog.Debug("saved response body", "id", r.reqBodyID)
	}
	r.timing.Stop()

	r.timing.Start(timing.TimeWriteResponse)
	err = r.resp.Write(tlsconn) // write the response to the TLS connection
	r.timing.Stop()
	if err != nil {
		return fmt.Errorf("tls connection write: %w", err)
	}

	return nil
}

// NOTE: handleNoMITM is falling out of support rn, gotta fix ts

// handleNoMITM handles an HTTPS connection without man-in-the-middling it. It just establishes a secure
// tunnel.
func (r *Request) handleNoMITM(m *Manager) error {
	// this code will need to be combined because it's the same in Perform and here in tunneling
	if config.DefaultConfig.RequireApproval {
		r.timing.Start(timing.TimeWaitApproval)
		if !m.RecieveApproval(r) { // req.Secure changes should not affect this (so we're good i think)
			return ErrPerformStop
		}
		r.timing.Stop()
	}
	if config.DefaultConfig.PerformDelay != 0 {
		r.timing.Start(timing.TimeDelayPerform)
		duration := time.Duration(config.DefaultConfig.PerformDelay) * time.Millisecond
		time.Sleep(duration)
		r.timing.Stop()
	}

	r.timing.Start(timing.TimeTunnel)
	defer r.timing.Stop()
	hconn, err := net.Dial("tcp", r.Host)
	if err != nil {
		return fmt.Errorf("dial host: %w", err)
	}

	// me gusta context :)
	ctx, cancel := context.WithCancel(context.Background())

	m.SendTunnel(r)
	go func() {
		defer cancel()
		_, err := io.Copy(hconn, r.conn)
		if err != nil {
			slog.Warn("io.Copy error (conn -> hostconn)", "err", err)
		}
	}()

	go func() {
		defer cancel()
		_, err := io.Copy(r.conn, hconn)
		if err != nil {
			slog.Warn("io.Copy error (hconn -> conn)", "err", err)
		}
	}()

	<-ctx.Done()

	return nil
}

// // hijack hijacks the ResponseWriter connection to the client.
// func hijack(w http.ResponseWriter) (net.Conn, error) {
// 	h, ok := w.(http.Hijacker)
// 	if !ok {
// 		return nil, ErrResponseWriterNoHijack
// 	}
// 	c, _, err := h.Hijack()
// 	return c, err
// }

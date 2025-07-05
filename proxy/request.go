package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	certificate "github.com/tiredkangaroo/bigproxy/proxy/certificates"
	"github.com/tiredkangaroo/bigproxy/proxy/http"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
	"github.com/tiredkangaroo/bigproxy/proxy/timing"
)

const ThisDevice string = "This Device"

var (
	ErrPerformStop = errors.New("perform stopped")
)

type Kind int64

const (
	RequestKindHTTP      Kind = iota // HTTP
	RequestKindHTTPS                 // HTTPS
	RequestKindHTTPSMITM             // HTTPS with MITM
)

func (k Kind) String() string {
	switch k {
	case RequestKindHTTP:
		return "HTTP (Insecure)"
	case RequestKindHTTPS:
		return "HTTPS (Secure)"
	case RequestKindHTTPSMITM:
		return "HTTPS (with MITM)"
	default:
		return "Unknown"
	}
}

func KindFromString(s string) Kind {
	switch s {
	case "HTTP (Insecure)":
		return RequestKindHTTP
	case "HTTPS (Secure)":
		return RequestKindHTTPS
	case "HTTPS (with MITM)":
		return RequestKindHTTPSMITM
	default:
		return -1 // Unknown kind
	}
}

type Request struct {
	// NOTE: most if not all of these fields should NOT be private fields
	Kind Kind // 0 = HTTP, 1 = HTTPS, 2 = HTTPS (with MITM)
	// this attribute should be renamed for its purpose
	Secure bool

	ID       string
	Starred  bool
	Datetime time.Time
	Host     string

	conn net.Conn

	ClientIP            string
	ClientPort          string
	ClientAuthorization string
	ClientProcessID     int
	ClientApplication   string

	timing *timing.Timing

	req        *http.Request
	reqBodyID  string // ID of the request body in the database
	resp       *http.Response
	respBodyID string // ID of the response body in the database

	errorText string // NOTE: only populated at db, prolly should change that, maybe not, who knows, not me, maybe me, who knows

	approveResponseFunc func(approved bool)
}

func (r *Request) Init(req *http.Request) error {
	r.timing.Start(timing.TimeRequestInit)
	defer r.timing.Stop()
	r.Secure = req.Method == http.MethodConnect
	r.reqBodyID = r.ID + "-req-body"
	r.respBodyID = r.ID + "-resp-body"

	// possibly being deprecated, idk
	if r.Secure && config.DefaultConfig.MITM {
		r.Kind = 2 // HTTPS with MITM
	} else if r.Secure {
		r.Kind = 1 // HTTPS
	} else {
		r.Kind = 0 // HTTP
	}

	r.Datetime = time.Now()

	r.Host = req.Host
	if _, _, err := net.SplitHostPort(req.Host); err != nil {
		if r.Secure {
			r.Host = req.Host + ":443" // default HTTPS port
		} else {
			r.Host = req.Host + ":80" // default HTTP port
		}
	}

	r.ClientIP = r.conn.RemoteAddr().String()
	if ip, port, err := net.SplitHostPort(r.ClientIP); err == nil {
		r.ClientIP = ip
		r.ClientPort = port
	}
	if ipIsLocalhost(r.ClientIP) {
		r.ClientIP = ThisDevice
	}

	r.ClientAuthorization = req.Header.Get("Proxy-Authorization")

	if config.DefaultConfig.GetClientProcessInfo {
		r.timing.Substart(timing.SubtimeGetClientProcessInfo)
		getClientProcessInfo(r.ClientIP, r.ClientPort, &r.ClientProcessID, &r.ClientApplication)
		r.timing.Substop()
	}

	return nil
}

// Perform performs the request and returns the raw response as a byte slice.
func (r *Request) Perform(m *Manager, c *certificate.Certificates) (*http.Response, error) {
	r.timing.Start(timing.TimePerformRequest)
	defer r.timing.Stop()
	r.req.Header.Del("Proxy-Authorization")
	r.req.Header.Del("Proxy-Connection")

	if config.DefaultConfig.RequireApproval {
		r.timing.Substart(timing.SubtimeWaitApproval)
		if !m.RecieveApproval(r) {
			return nil, ErrPerformStop
		}
		r.timing.Substop()
	}

	if config.DefaultConfig.PerformDelay != 0 {
		r.timing.Substart(timing.SubtimeDelayPerform)
		duration := time.Duration(config.DefaultConfig.PerformDelay) * time.Millisecond
		time.Sleep(duration)
		r.timing.Substop()
	}

	r.timing.Substart(timing.SubtimeDialHost)
	var hostconn net.Conn
	var err error
	if r.Secure {
		var sysCertPool *x509.CertPool
		sysCertPool, err = c.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("get system cert pool: %w", err)
		}
		hostconn, err = tls.Dial("tcp", r.Host, &tls.Config{
			RootCAs: sysCertPool,
		})
	} else {
		hostconn, err = net.Dial("tcp", r.Host)
	}
	if err != nil {
		return nil, fmt.Errorf("dial host: %w", err)
	}
	r.timing.Substop()

	r.timing.Substart(timing.SubtimeWriteRequest)
	if err := r.req.Write(hostconn); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}
	r.timing.Substop()

	r.timing.Substart(timing.SubtimeReadResponse)
	resp, err := http.ReadResponse(hostconn)
	r.timing.Substop()
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	r.resp = resp
	return r.resp, nil
}

func (r *Request) BytesTransferred() int64 {
	cc := r.conn.(*CustomConn)
	return cc.BytesTransferred()
}

func (r *Request) MarshalJSON() ([]byte, error) {
	var state string
	if r.errorText != "" {
		state = "Error"
	} else {
		state = "Done"
	}
	return json.Marshal(map[string]any{
		"id":                  r.ID,
		"starred":             r.Starred,
		"datetime":            r.Datetime.UnixMilli(), // unix milli for js
		"secure":              r.Secure,
		"clientIP":            r.ClientIP,
		"clientApplication":   r.ClientApplication,
		"clientAuthorization": r.ClientAuthorization,
		"host":                r.Host,

		"method":     r.req.Method.String(),
		"path":       r.req.Path,
		"query":      r.req.Query,
		"headers":    r.req.Header,
		"bodyID":     r.reqBodyID,
		"bodyLength": r.req.ContentLength,

		"response": map[string]any{
			"statusCode": r.resp.StatusCode,
			"headers":    r.resp.Header,
			"bodyID":     r.respBodyID,
			"bodyLength": r.resp.ContentLength,
		},

		"state":        state,
		"error":        r.errorText,
		"timing":       r.timing.Export(),
		"timing_total": r.timing.Total(),
	})
}

// func dial(host string) (net.Conn, error) {}

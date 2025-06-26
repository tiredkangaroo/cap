package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/tiredkangaroo/bigproxy/proxy/http"

	"github.com/google/uuid"
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

func (r *Request) Init(conn net.Conn, req *http.Request) error {
	r.Secure = req.Method == http.MethodConnect
	r.timing = timing.New()

	// possibly being deprecated, idk
	if r.Secure && config.DefaultConfig.MITM {
		r.Kind = 2 // HTTPS with MITM
	} else if r.Secure {
		r.Kind = 1 // HTTPS
	} else {
		r.Kind = 0 // HTTP
	}

	r.timing.Start(timing.TimeRequestInit)
	defer r.timing.Stop()

	r.Datetime = time.Now()

	r.timing.Substart(timing.SubtimeUUID)
	id, err := uuid.NewRandom()
	if err != nil {
		slog.Error("uuid error", "err", err.Error())
		r.ID = "75756964-7634-6765-6e65-72726f720000" // this isn't a random UUID
	}
	r.ID = id.String()
	r.reqBodyID = r.ID + "-req-body"
	r.respBodyID = r.ID + "-resp-body"
	r.timing.Substop()

	r.conn = &CustomConn{
		u: conn,
	}

	r.Host = req.Host

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
func (r *Request) Perform(m *Manager) (*http.Response, error) {
	// NOTE: these should be sub times where Perform is the major time
	r.timing.Start(timing.TimePrepRequest)

	r.req.Header.Del("Proxy-Authorization")
	r.req.Header.Del("Proxy-Connection")

	// NOTE: this is temporarily unsupported while we make some changes
	// if config.DefaultConfig.RealIPHeader {
	// 	// r.req.Header.Set("X-Forwarded-For", r.req.RemoteAddr)
	// }

	r.timing.Stop()
	if config.DefaultConfig.RequireApproval {
		r.timing.Start(timing.TimeWaitApproval)
		if !m.RecieveApproval(r) {
			return nil, ErrPerformStop
		}
		r.timing.Stop()
	}

	if config.DefaultConfig.PerformDelay != 0 {
		r.timing.Start(timing.TimeDelayPeform)
		duration := time.Duration(config.DefaultConfig.PerformDelay) * time.Millisecond
		time.Sleep(duration)
		r.timing.Stop()
	}

	r.timing.Start(timing.TimeDialHost)
	fmt.Println("dialing host", r.req.Host)
	hostc, err := net.Dial("tcp", r.req.Host)
	if err != nil {
		return nil, fmt.Errorf("dial host: %w", err)
	}
	r.timing.Stop()
	fmt.Println("dialed host", r.req.Host)

	fmt.Println("writing request to host", r.req.Host)
	r.timing.Start(timing.TimeWriteRequest)
	if err := r.req.Write(hostc); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}
	r.timing.Stop()
	fmt.Println("wrote request to host", r.req.Host)

	fmt.Println("reading response")
	return http.ReadResponse(hostc)
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
		"datetime":            r.Datetime.UnixMilli(), // unix milli for js
		"secure":              r.Secure,
		"clientIP":            r.ClientIP,
		"clientApplication":   r.ClientApplication,
		"clientAuthorization": r.ClientAuthorization,
		"host":                r.Host,

		"method":  r.req.Method.String(),
		"path":    r.req.Path,
		"query":   r.req.Query,
		"headers": r.req.Header,

		"response": map[string]any{
			"statusCode": r.resp.StatusCode,
			"headers":    r.resp.Header,
		},

		"state":        state,
		"error":        r.errorText,
		"timing":       r.timing.Export(),
		"timing_total": r.timing.Total(),
	})
}

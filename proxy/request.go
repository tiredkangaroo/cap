package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

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

	req     *http.Request
	reqBody []byte
	resp    *http.Response

	errorText string // NOTE: only populated at db, prolly should change that

	approveResponseFunc func(approved bool)
}

func (r *Request) Init(w http.ResponseWriter, req *http.Request) error {
	// NOTE: mitm config can change in the middle of the request (it shouldn't be able to change behavior of the request after init)
	r.req = req
	r.Secure = r.req.Method == http.MethodConnect
	r.timing = timing.New(r.Secure, config.DefaultConfig.MITM)

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
	r.timing.Substop()

	// hijack the connection
	conn, err := hijack(w)
	if err != nil {
		return fmt.Errorf("hijack error: %w", err)
	}
	r.conn = &CustomConn{
		u: conn,
	}

	r.Host = r.req.Host

	r.ClientIP = r.req.RemoteAddr
	if ip, port, err := net.SplitHostPort(r.ClientIP); err == nil {
		r.ClientIP = ip
		r.ClientPort = port
	}
	if ipIsLocalhost(r.ClientIP) {
		r.ClientIP = ThisDevice
	}

	r.ClientAuthorization = r.req.Header.Get("Proxy-Authorization")

	if config.DefaultConfig.GetClientProcessInfo {
		r.timing.Substart(timing.SubtimeGetClientProcessInfo)
		getClientProcessInfo(r.ClientIP, r.ClientPort, &r.ClientProcessID, &r.ClientApplication)
		r.timing.Substop()
	}

	return nil
}

// Perform performs the request and returns the raw response as a byte slice.
func (r *Request) Perform(m *Manager) (*http.Response, []byte, error) {
	// might be too resource heavy to do it this way
	r.timing.Start(timing.TimePrepRequest)
	// toURL is used to convert the host to a valid URL.
	newURL, err := toURL(r.req.Host, r.Secure)
	if err != nil {
		return nil, nil, fmt.Errorf("malformed url (toURL): %w", err)
	}
	newURL.Path = r.req.URL.Path
	newURL.RawQuery = r.req.URL.RawQuery

	// set the NewURL and remove unnecessary headers from the request
	r.req.URL = newURL
	r.req.RequestURI = "" // clears the request URI to avoid error
	r.req.Header.Del("Proxy-Authorization")
	r.req.Header.Del("Proxy-Connection")

	if config.DefaultConfig.RealIPHeader {
		r.req.Header.Set("X-Forwarded-For", r.req.RemoteAddr)
	}

	r.timing.Stop()
	if config.DefaultConfig.RequireApproval {
		r.timing.Start(timing.TimeWaitApproval)
		if !m.RecieveApproval(r) {
			return nil, nil, ErrPerformStop
		}
		r.timing.Stop()
	}

	if config.DefaultConfig.PerformDelay != 0 {
		r.timing.Start(timing.TimeDelayPeform)
		duration := time.Duration(config.DefaultConfig.PerformDelay) * time.Millisecond
		time.Sleep(duration)
		r.timing.Stop()
	}

	r.timing.Start(timing.TimeRequestPerform)
	// do the request
	resp, err := http.DefaultClient.Do(r.req)
	r.timing.Stop()
	if err != nil {
		return nil, nil, fmt.Errorf("req do error: %w", err)
	}

	r.timing.Start(timing.TimeDumpResponse)
	data, err := httputil.DumpResponse(resp, true)
	r.timing.Stop()
	if err != nil {
		return nil, nil, fmt.Errorf("dump server response: %w", err)
	}
	return resp, data, nil
}

func (r *Request) BytesTransferred() int64 {
	cc := r.conn.(*CustomConn)
	return cc.BytesTransferred()
}

// body reads the body of the request and returns it as a byte slice. If the request
// is an HTTPS tunnel request, the body will return a message indicating that the configuration
// does not allow for the body to be read. If the body cannot be read for any another reason, it returns a
// nil byte slice.
func (r *Request) body() []byte {
	if r.reqBody != nil {
		// if the body is already read, return it
		return r.reqBody
	}
	// this is resource heavy, but if we're doing it, might as well just use sync.Once
	var bodyData []byte
	var err error
	if config.DefaultConfig.ProvideRequestBody {
		bodyData, err = io.ReadAll(r.req.Body)
		if err != nil {
			slog.Error("body data read", "err", err.Error())
			bodyData = nil
		}
		r.req.Body = io.NopCloser(bytes.NewBuffer(bodyData)) // make sure the body can be read again
		r.reqBody = bodyData                                 // store the body data for later use
	} else {
		bodyData = nil
	}
	return bodyData
}

func (r *Request) respbody() []byte {
	var bodyData []byte
	var err error
	// FIXME: logic is too complex
	if config.DefaultConfig.ProvideResponseBody {
		var rd = r.resp.Body
		if r.resp.Header.Get("Content-Encoding") == "gzip" {
			rd, err = gzip.NewReader(r.resp.Body)
			if err != nil {
				// NOTE: possible bug here
				// slog.Error("gzip reader error", "err", err.Error())
				rd = r.resp.Body // fallback to the original body if gzip reader fails
			} else {
				defer rd.Close() // ensure the gzip reader is closed after reading
			}
		}
		bodyData, err = io.ReadAll(rd)
		if err != nil {
			slog.Error("body data read", "err", err.Error())
			bodyData = nil
		}
		r.resp.Body = io.NopCloser(bytes.NewBuffer(bodyData)) // make sure the body can be read again
	} else {
		bodyData = nil
	}
	return bodyData
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
		"secureState":         r.Kind.String(),
		"clientIP":            r.ClientIP,
		"clientApplication":   r.ClientApplication,
		"clientAuthorization": r.ClientAuthorization,
		"host":                r.Host,

		"url":     r.req.URL.String(), // NOTE: not included in frontend's types.ts, but it will be useful (maybe, maybe not, who is to tell)
		"method":  r.req.Method,
		"path":    r.req.URL.Path,
		"query":   r.req.URL.Query(),
		"headers": r.req.Header,
		"body":    string(r.body()),

		"response": map[string]any{
			"statusCode": r.resp.StatusCode,
			"headers":    r.resp.Header,
			"body":       string(r.respbody()),
		},

		"state":        state,
		"error":        r.errorText,
		"timing":       r.timing.Export(),
		"timing_total": r.timing.Total(),
	})
}

package main

import (
	"bytes"
	"compress/gzip"
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
)

var (
	ErrPerformStop = errors.New("perform stopped")
)

type Request struct {
	id                  string
	datetime            time.Time
	host                string
	conn                net.Conn
	secure              bool
	clientIP            string
	clientAuthorization string

	req  *http.Request
	resp *http.Response

	approveResponseFunc func(approved bool)
}

func (r *Request) Init(w http.ResponseWriter, req *http.Request) error {
	r.datetime = time.Now()
	id, err := uuid.NewRandom()
	if err != nil {
		slog.Error("uuid error", "err", err.Error())
		r.id = "75756964-7634-6765-6e65-72726f720000" // this isn't a random UUID
	}
	r.id = id.String()

	// hijack the connection
	conn, err := hijack(w)
	if err != nil {
		return fmt.Errorf("hijack error: %w", err)
	}
	r.conn = conn

	r.req = req
	r.secure = r.req.Method == http.MethodConnect
	r.host = r.req.Host
	r.clientIP = r.req.RemoteAddr
	r.clientAuthorization = r.req.Header.Get("Proxy-Authorization")

	return nil
}

// Perform performs the request and returns the raw response as a byte slice.
func (r *Request) Perform(m *Manager) (*http.Response, []byte, error) {
	// might be too resource heavy to do it this way

	// toURL is used to convert the host to a valid URL.
	newURL, err := toURL(r.req.Host, r.secure)
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

	if config.DefaultConfig.RequireApproval {
		if !m.RecieveApproval(r) {
			return nil, nil, ErrPerformStop
		}
	}
	if config.DefaultConfig.PerformDelay != 0 {
		time.Sleep(time.Duration(config.DefaultConfig.PerformDelay) * time.Millisecond)
	}

	// do the request
	resp, err := http.DefaultClient.Do(r.req)
	if err != nil {
		return nil, nil, fmt.Errorf("req do error: %w", err)
	}

	data, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, nil, fmt.Errorf("dump server response: %w", err)
	}
	return resp, data, nil
}

// body reads the body of the request and returns it as a byte slice. If the request
// is an HTTPS tunnel request, the body will return a message indicating that the configuration
// does not allow for the body to be read. If the body cannot be read for any another reason, it returns a
// nil byte slice.
func (r *Request) body() []byte {
	var bodyData []byte
	var err error
	if config.DefaultConfig.ProvideRequestBody {
		bodyData, err = io.ReadAll(r.req.Body)
		if err != nil {
			slog.Error("body data read", "err", err.Error())
			bodyData = nil
		}
		r.req.Body = io.NopCloser(bytes.NewBuffer(bodyData)) // make sure the body can be read again
	} else {
		bodyData = []byte("body will not be provided under configuration rules")
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
				slog.Error("gzip reader error", "err", err.Error())
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
		bodyData = []byte("body will not be provided under configuration rules")
	}
	return bodyData
}

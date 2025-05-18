package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/google/uuid"
)

type Request struct {
	id   string
	date time.Time

	secure bool

	req  *http.Request
	host string

	conn                net.Conn
	clientIP            string
	clientAuthorization string
}

func (r *Request) Init(w http.ResponseWriter, req *http.Request) error {
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

func (r *Request) Perform(config *Config) ([]byte, error) {
	// toURL is used to convert the host to a valid URL.
	newURL, err := toURL(r.req.Host, r.secure)
	if err != nil {
		return nil, fmt.Errorf("malformed url (toURL): %w", err)
	}
	newURL.Path = r.req.URL.Path
	newURL.RawQuery = r.req.URL.RawQuery

	// set the NewURL and remove unnecessary headers from the request
	r.req.URL = newURL
	r.req.RequestURI = "" // clears the request URI to avoid error
	r.req.Header.Del("Proxy-Authorization")
	r.req.Header.Del("Proxy-Connection")

	if config.RealIPHeader {
		r.req.Header.Set("X-Forwarded-For", r.req.RemoteAddr)
	}

	// do the request
	resp, err := http.DefaultClient.Do(r.req)
	if err != nil {
		return nil, fmt.Errorf("req do error: %w", err)
	}

	// might be too resource heavy to do it this way
	data, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, fmt.Errorf("dump server response: %w", err)
	}
	return data, nil
}

// followUpDataWithMITMInfo provides data to be sent to the client that has
func (r *Request) followUpDataWithMITMInfo(config *Config) []byte {
	var bodyData []byte
	var err error
	if config.ProvideRequestBody {
		bodyData, err = io.ReadAll(r.req.Body)
		if err != nil {
			bodyData = ResponseReadInterceptedBodyError
			slog.Error("body data read", "err", err.Error())
		}
		r.req.Body = io.NopCloser(bytes.NewBuffer(bodyData)) // make sure the body can be read again
	} else {
		bodyData = []byte("body will not be provided under configuration rules")
	}
	data, _ := json.Marshal(map[string]any{
		"id":      r.id,
		"method":  r.req.Method,
		"headers": r.req.Header,
		"body":    string(bodyData),
	})
	return data
}

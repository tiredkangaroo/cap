package http

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/tiredkangaroo/cap/proxy/config"
)

var (
	ErrProtocolError              = errors.New("protocol error: invalid HTTP request/response")
	ErrMissingHostHeader          = errors.New("missing Host header in HTTP request")
	ErrContentLengthInvalid       = errors.New("invalid Content-Length header in HTTP request")
	ErrMissingContentLengthHeader = errors.New("missing Content-Length header in HTTP request")
)

// we can do pooling of http.Request later

func ReadRequest(conn net.Conn) (*Request, error) {
	// example request:
	// POST /cgi-bin/process.cgi HTTP/1.1
	// User-Agent: Mozilla/4.0 (compatible; MSIE5.01; Windows NT)
	// Host: www.tutorialspoint.com
	// Content-Type: application/x-www-form-urlencoded
	// Content-Length: length
	// Accept-Language: en-us
	// Accept-Encoding: gzip, deflate
	// Connection: Keep-Alive

	// licenseID=string&content=string&/paramsXML=string
	req := NewRequest()
	req.conn = conn

	buf := bufio.NewReader(conn)
	data, err := buf.ReadBytes('\n')
	if err != nil {
		if config.DefaultConfig.Debug {
			slog.Error("http parser: failed to read first line", "err", err.Error())
		}
		return nil, err
	}
	firstLineData := bytes.Split(data, []byte{' '})
	if len(firstLineData) < 3 {
		if config.DefaultConfig.Debug {
			slog.Error("http parser: invalid first request line", "line", b2s(data))
		}
		return nil, ErrProtocolError
	}
	req.Method = methodFromBytes(firstLineData[0])
	if req.Method == MethodUnknown {
		if config.DefaultConfig.Debug {
			slog.Error("http parser: invalid method", "method", b2s(firstLineData[0]))
		}
		return nil, ErrProtocolError
	}
	u, err := url.Parse(b2s(firstLineData[1]))
	if err != nil {
		if config.DefaultConfig.Debug {
			slog.Error("http parser: failed to parse path", "url", b2s(firstLineData[1]), "err", err.Error())
		}
		return nil, fmt.Errorf("failed to parse path: %w", err)
	}
	req.Path = u.Path
	req.Query = u.Query()

	req.Proto = bytes.TrimSpace(firstLineData[2])

	req.Header, err = readHeader(buf)
	if err != nil {
		return nil, err
	}
	if err := manageSpecialRequestHeaders(req); err != nil {
		return nil, fmt.Errorf("special headers issue: %w", err)
	}

	req.Body = NewBody(buf, req.ContentLength)
	if req.ContentLength == 0 {
		// if Content-Length is 0, we don't need to read the body. this releases the buffer.
		req.Body.buf = nil
	}

	return req, nil
}

func manageSpecialRequestHeaders(req *Request) error {
	req.Host = req.Header.Get("Host")
	if req.Host == "" {
		return ErrMissingHostHeader
	}

	contentLength := req.Header.Get("Content-Length")
	if (req.Method == MethodPost || req.Method == MethodPut || req.Method == MethodPatch) && contentLength == "" {
		return ErrMissingContentLengthHeader
	}
	if contentLength != "" {
		cl, err := strconv.Atoi(contentLength)
		if err != nil {
			return ErrContentLengthInvalid
		}
		req.ContentLength = int64(cl)
	}

	req.Connection = req.Header.Get("Connection")

	return nil
}

// returns new path, query values
func parseQuery(path string) (string, url.Values) {
	queryI := strings.IndexByte(path, '?')
	if queryI == -1 {
		return path, url.Values{} // no query
	}
	queryStr := path[queryI+1:]
	q, err := url.ParseQuery(queryStr)
	if err != nil {
		return path, url.Values{}
	}
	return path[:queryI], q
}

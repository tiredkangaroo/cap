package http

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"strings"
	"unsafe"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
)

var (
	ErrProtocolError              = errors.New("protocol error: invalid HTTP request")
	ErrMissingHostHeader          = errors.New("missing Host header in HTTP request")
	ErrContentLengthInvalid       = errors.New("invalid Content-Length header in HTTP request")
	ErrMissingContentLengthHeader = errors.New("missing Content-Length header in HTTP request")
)

// we can do pooling of http.Request later

func ReadRequest(conn net.Conn) (*Request, error) {
	var rnSuffix = []byte("\r\n")
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
		return nil, ErrProtocolError
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
	req.Path = b2s(firstLineData[1])
	// the query is everything after the first "?" in the path
	req.Path, req.Query = parseQuery(req.Path)

	req.Proto = bytes.TrimSpace(firstLineData[2])

	req.Header = make(map[string][]string)
	// len(data) > 2 is to ensure the thing we just read isn't \r\n (indicates the end of headers)
	for data, err = buf.ReadBytes('\n'); err == nil && len(data) > 2; data, err = buf.ReadBytes('\n') {
		// split by ": " to get key and value
		keyVSplit := bytes.SplitN(data, []byte{':', ' '}, 2)
		if len(keyVSplit) != 2 {
			if config.DefaultConfig.Debug {
				slog.Error("http parser: invalid header line", "line", b2s(data))
			}
			return nil, ErrProtocolError
		}
		key := b2s(keyVSplit[0])
		value := b2s(bytes.TrimSuffix(keyVSplit[1], rnSuffix))
		if len(key) == 0 {
			if config.DefaultConfig.Debug {
				slog.Error("http parser: empty header key", "line", b2s(data))
			}
			return nil, ErrProtocolError
		}
		if req.Header[key] == nil {
			req.Header[key] = []string{value}
		} else {
			req.Header[key] = append(req.Header[key], value)
		}
		if err := handleSpecialHeaders(req, key, value); err != nil {
			if config.DefaultConfig.Debug {
				slog.Error("http parser: error handling special header", "key", key, "value", value, "err", err.Error())
			}
			return nil, err
		}
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	if err := ensureSpecialHeaders(req); err != nil {
		if config.DefaultConfig.Debug {
			slog.Error("http parser: missing special headers", "err", err.Error())
		}
		return nil, err
	}

	buf.Reset(nil)
	buf.Reset(conn)

	req.Body = NewBody(buf, req.ContentLength)
	if req.ContentLength == 0 {
		// if Content-Length is 0, we don't need to read the body. this releases the buffer.
		req.Body.buf = nil
	}

	return req, nil
}

func handleSpecialHeaders(req *Request, key, value string) error {
	switch key {
	case "Host":
		req.Host = value
	case "Content-Length":
		cl, err := strconv.Atoi(value)
		if err != nil {
			if config.DefaultConfig.Debug {
				slog.Error("http parser: invalid Content-Length header", "value", value, "err", err.Error())
			}
			return ErrContentLengthInvalid
		}
		req.ContentLength = int64(cl)
	case "Connection":
		req.Connection = value
	}
	return nil
}

func ensureSpecialHeaders(req *Request) error {
	if req.Host == "" {
		return ErrMissingHostHeader
	}
	if req.ContentLength < 0 {
		if config.DefaultConfig.Debug {
			slog.Error("http parser: missing Content-Length header")
		}
		return ErrMissingContentLengthHeader
	}
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

func b2s(b []byte) string {
	return unsafe.String(&b[0], len(b))
}
func s2b(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

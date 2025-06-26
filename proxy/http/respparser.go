package http

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
)

var (
	ErrResponseLineInvalid = errors.New("invalid response line")
	ErrInvalidStatusCode   = errors.New("invalid status code in response")
)

func ReadResponse(conn net.Conn) (*Response, error) {
	// HTTP/1.1 200 OK
	// Date: Tue, 22 Jun 2024 16:00:00 GMT
	// Content-Type: text/html; charset=UTF-8
	// Content-Length: 1234

	// <!DOCTYPE html>
	// <html>
	// <head>
	// <title>Example Page</title>
	// </head>
	// <body>
	// <h1>Hello, World!</h1>
	// </body>
	// </html>

	buf := bufio.NewReader(conn)
	resp := NewResponse()

	conn.SetReadDeadline(time.Now().Add(time.Minute))
	fmt.Println("reading first line")
	data, err := buf.ReadBytes('\n')
	if err != nil {
		if config.DefaultConfig.Debug {
			slog.Error("http parser: failed to read first line", "err", err.Error())
		}
		return nil, ErrProtocolError
	}
	fmt.Println("read first line", b2s(data))
	firstLineData := bytes.Split(data, []byte{' '})
	if len(firstLineData) < 3 {
		if config.DefaultConfig.Debug {
			slog.Error("http parser: invalid first request line", "line", b2s(data))
		}
		return nil, ErrResponseLineInvalid
	}

	resp.Version = firstLineData[0]

	resp.StatusCode, err = strconv.Atoi(b2s(firstLineData[1]))
	if err != nil {
		if config.DefaultConfig.Debug {
			slog.Error("http parser: invalid status code", "status", b2s(firstLineData[1]), "err", err.Error())
		}
		return nil, ErrInvalidStatusCode
	}
	fmt.Println("pared data status code", resp.StatusCode)
	if statusString(resp.StatusCode) == statusString(StatusUnknown) {
		if config.DefaultConfig.Debug {
			slog.Error("http parser: invalid status code", "status", b2s(firstLineData[1]))
		}
		return nil, ErrInvalidStatusCode
	}

	// NOTE: status string will not be handled for now

	fmt.Println("reading headers")
	resp.Header, err = readHeader(buf)
	if err != nil {
		return nil, err
	}
	fmt.Println("read headers")
	if err := manageSpecialResponseHeaders(resp); err != nil {
		return nil, fmt.Errorf("special headers issue: %w", err)
	}

	conn.SetDeadline(time.Time{})

	fmt.Println("79", resp.Header, resp.ContentLength)
	resp.Body = NewBody(buf, resp.ContentLength)
	if resp.ContentLength == 0 {
		resp.Body.buf = nil // release at once
	}

	return resp, nil
}

func manageSpecialResponseHeaders(resp *Response) error {
	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		cl, err := strconv.Atoi(contentLength)
		if err != nil {
			return fmt.Errorf("invalid Content-Length header: %w", err)
		}
		resp.ContentLength = int64(cl)
	}
	return nil
}

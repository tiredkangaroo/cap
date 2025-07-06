package http_test

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/tiredkangaroo/cap/proxy/http"
)

type MockNetConn struct {
	*bufio.ReadWriter
}

func (m *MockNetConn) Close() error {
	return nil
}
func (m *MockNetConn) LocalAddr() net.Addr {
	return nil
}
func (m *MockNetConn) RemoteAddr() net.Addr {
	return nil
}
func (m *MockNetConn) SetDeadline(t time.Time) error {
	return nil
}
func (m *MockNetConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (m *MockNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestReadRequest(t *testing.T) {
	c := NewMockNetConn()
	ex := `GET /endpoint?q=1&world=thecraziestworld HTTP/1.1\r\n` +
		`Host: example.com\r\n` +
		`User-Agent: Go-http-client/1.1\r\n` +
		`Accept: */*\r\n` +
		`Accept-Encoding: gzip\r\n` +
		`Connection: keep-alive\r\n` +
		`Content-Type: text/plain\r\n` +
		`Content-Length: 14\r\n` +
		`\r\n` +
		`char1234567890`
	_, err := c.Write([]byte(ex))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	req, err := http.ReadRequest(c)
	if err != nil {
		t.Fatalf("ReadRequest failed: %v", err)
	}
	if req == nil {
		t.Fatal("expected non-nil request, got nil")
	}
	if string(req.Proto) != "HTTP/1.1" {
		t.Errorf("expected HTTP version HTTP/1.1, got %s", req.Proto)
	}
	if req.Method != http.MethodGet {
		t.Errorf("expected method %s, got %s", http.MethodGet, req.Method)
	}
	if req.Path != "/endpoint" {
		t.Errorf("expected path %s, got %s", "/endpoint", req.Path)
	}
	if req.Query.Get("q") != "1" {
		t.Errorf("expected query q=1, got %s", req.Query.Get("q"))
	}
	if req.Query.Get("world") != "thecraziestworld" {
		t.Errorf("expected query world=thecraziestworld, got %s", req.Query.Get("world"))
	}
	if req.Header.Get("Host") != "example.com" {
		t.Errorf("expected Host header example.com, got %s", req.Header.Get("Host"))
	}
	if req.Header.Get("User-Agent") != "Go-http-client/1.1" {
		t.Errorf("expected User-Agent header Go-http-client/1.1, got %s", req.Header.Get("User-Agent"))
	}
	if req.Header.Get("Accept") != "*/*" {
		t.Errorf("expected Accept header */*, got %s", req.Header.Get("Accept"))
	}
	if req.Header.Get("Accept-Encoding") != "gzip" {
		t.Errorf("expected Accept-Encoding header gzip, got %s", req.Header.Get("Accept-Encoding"))
	}
	if req.Header.Get("Connection") != "keep-alive" {
		t.Errorf("expected Connection header keep-alive, got %s", req.Header.Get("Connection"))
	}
	if req.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type header text/plain, got %s", req.Header.Get("Content-Type"))
	}
	if req.Header.Get("Content-Length") != "14" {
		t.Errorf("expected Content-Length header 14, got %s", req.Header.Get("Content-Length"))
	}
	if req.Body == nil {
		t.Error("expected Body to be initialized, got nil")
	}
	if req.Body.ContentLength() != 14 {
		t.Errorf("expected Body length 14, got %d", req.Body.ContentLength())
	}
	b, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if string(b) != "char1234567890" {
		t.Errorf("expected body 'char1234567890', got '%s'", string(b))
	}
}

func NewMockNetConn() *MockNetConn {
	b := bytes.NewBuffer(make([]byte, 0, 512))
	return &MockNetConn{
		ReadWriter: bufio.NewReadWriter(bufio.NewReader(b), bufio.NewWriter(b)),
	}
}

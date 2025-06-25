package http

import (
	"io"
	"net"
	"net/url"
)

type Method uint8

const (
	MethodUnknown Method = iota
	MethodGet
	MethodPost
	MethodPut
	MethodPatch
	MethodDelete
	MethodOptions
	MethodHead
	MethodConnect
	MethodTrace
)

func MethodFromString(s string) Method {
	switch s {
	case "GET":
		return MethodGet
	case "POST":
		return MethodPost
	case "PUT":
		return MethodPut
	case "PATCH":
		return MethodPatch
	case "DELETE":
		return MethodDelete
	case "OPTIONS":
		return MethodOptions
	case "HEAD":
		return MethodHead
	case "CONNECT":
		return MethodConnect
	case "TRACE":
		return MethodTrace
	default:
		return MethodUnknown
	}
}

// type Connection uint8

// const (
// 	ConnectionUnknown Connection = iota
// 	ConnectionKeepAlive
// 	ConnectionClose
// 	ConnectionUpgrade
// )

func (m Method) String() string {
	switch m {
	case MethodGet:
		return "GET"
	case MethodPost:
		return "POST"
	case MethodPut:
		return "PUT"
	case MethodPatch:
		return "PATCH"
	case MethodDelete:
		return "DELETE"
	case MethodOptions:
		return "OPTIONS"
	case MethodHead:
		return "HEAD"
	case MethodConnect:
		return "CONNECT"
	case MethodTrace:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

type Request struct {
	conn net.Conn

	Method Method
	Path   string
	Query  url.Values
	Header Header

	Host          string
	ContentLength int64
	Connection    string

	Proto []byte // HTTP version, e.g., "HTTP/1.1"

	Body *Body
}

// Write writes the HTTP request to the provided writer. It writes the request line, headers, and body if present.
func (r *Request) Write(w io.Writer) error {
	if r.conn == nil {
		return io.ErrClosedPipe
	}

	// request line
	var requestLine = make([]byte, 0, 64)
	requestLine = append(requestLine, s2b(r.Method.String())...)
	requestLine = append(requestLine, ' ')
	requestLine = append(requestLine, s2b(r.Path)...)
	requestLine = append(requestLine, ' ')
	requestLine = append(requestLine, r.Proto...)
	requestLine = append(requestLine, '\r', '\n')
	_, err := w.Write(requestLine)
	if err != nil {
		return err
	}

	// headers
	for key, values := range r.Header {
		for _, value := range values {
			headerLine := make([]byte, 0, len(key)+len(value)+4)
			headerLine = append(headerLine, s2b(key)...)   // len key
			headerLine = append(headerLine, ": "...)       // 2
			headerLine = append(headerLine, s2b(value)...) // len value
			headerLine = append(headerLine, '\r', '\n')    // 2
			_, err = w.Write(headerLine)
			if err != nil {
				return err
			}
		}
	}
	// end of headers
	_, err = w.Write([]byte{'\r', '\n'})
	if err != nil {
		return err
	}

	// body
	if r.Body != nil {
		_, err = r.Body.WriteTo(w)
		return err
	}
	return nil
}

func (r *Request) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func NewRequest() *Request {
	return &Request{
		Method:        MethodUnknown,
		Path:          "",
		Header:        make(map[string][]string),
		Host:          "",
		ContentLength: -1, // -1 indicates unknown content length
		Connection:    "",
	}
}

func methodFromBytes(b []byte) Method {
	if len(b) < 3 { // too short to be a valid method
		return MethodUnknown
	}
	switch b[0] {
	case 'G': // GET
		if len(b) != 3 {
			return MethodUnknown
		}
		return MethodGet
	case 'P':
		switch b[1] {
		case 'O': // POST
			if len(b) != 4 {
				return MethodUnknown
			}
			return MethodPost
		case 'U': // PUT
			if len(b) != 3 {
				return MethodUnknown
			}
			return MethodPut
		case 'A': // PATCH
			if len(b) != 5 {
				return MethodUnknown
			}
			return MethodPatch
		default:
			return MethodUnknown
		}
	case 'D': // DELETE
		if len(b) != 6 {
			return MethodUnknown
		}
		return MethodDelete
	case 'H': // HEAD
		if len(b) != 4 {
			return MethodUnknown
		}
		return MethodHead
	case 'C': // CONNECT
		if len(b) != 7 {
			return MethodUnknown
		}
		return MethodConnect
	case 'O': // OPTIONS
		if len(b) != 7 {
			return MethodUnknown
		}
		return MethodOptions
	case 'T': // TRACE
		if len(b) != 5 {
			return MethodUnknown
		}
		return MethodTrace
	default:
		return MethodUnknown
	}
}

// func connectionFromString(c string) Connection {
// 	switch strings.ToLower(c) {
// 	case "keep-alive":
// 		return ConnectionKeepAlive
// 	case "close":
// 		return ConnectionClose
// 	case "upgrade":
// 		return ConnectionUpgrade
// 	default:
// 		return ConnectionUnknown
// 	}
// }

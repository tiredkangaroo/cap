package http

import (
	"net"
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

	Method  Method
	Path    string
	Headers map[string][]string

	Host          string
	ContentLength int
	Connection    string

	Body *Body
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
		Headers:       make(map[string][]string),
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

package http

import (
	"bufio"
	"bytes"
	"log/slog"
	"net/textproto"
	"net/url"
	"unsafe"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
)

func readHeader(buf *bufio.Reader) (Header, error) {
	rnSuffix := []byte("\r\n")
	header := make(map[string][]string)
	// len(data) > 2 is to ensure the thing we just read isn't \r\n (indicates the end of headers)
	for data, err := buf.ReadBytes('\n'); err == nil && len(data) > 2; data, err = buf.ReadBytes('\n') {
		// split by ": " to get key and value
		keyVSplit := bytes.SplitN(data, []byte{':', ' '}, 2)
		if len(keyVSplit) != 2 {
			if config.DefaultConfig.Debug {
				slog.Error("http parser: invalid header line", "line", b2s(data))
			}
			return nil, ErrProtocolError
		}
		key := textproto.CanonicalMIMEHeaderKey(b2s(keyVSplit[0]))
		value := b2s(bytes.TrimSuffix(keyVSplit[1], rnSuffix))
		if len(key) == 0 {
			if config.DefaultConfig.Debug {
				slog.Error("http parser: empty header key", "line", b2s(data))
			}
			return nil, ErrProtocolError
		}
		if header[key] == nil {
			header[key] = []string{value}
		} else {
			header[key] = append(header[key], value)
		}
	}
	return header, nil
}

func b2s(b []byte) string {
	return unsafe.String(&b[0], len(b))
}
func s2b(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func queryString(q url.Values) []byte {
	if len(q) == 0 {
		return []byte{}
	}
	return s2b("?" + q.Encode())
}

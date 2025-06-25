package http

import (
	"bufio"
	"io"
	"log/slog"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
)

type Body struct {
	buf           *bufio.Reader
	readN         int64
	contentLength int64
}

func (buf *Body) Read(p []byte) (n int, err error) {
	if buf.readN >= buf.contentLength || buf.buf == nil {
		buf.buf = nil // release the buffer (we're done reading)
		return 0, io.EOF
	}

	n, err = buf.buf.Read(p)
	if err != nil && err != io.EOF {
		return n, err
	}
	maxRead := int(buf.contentLength - buf.readN)
	if n > maxRead {
		n = maxRead
		if config.DefaultConfig.Debug {
			slog.Warn("http body: read more than content length", "readN", buf.readN, "contentLength", buf.contentLength, "n", n)
		}
		clear(p[n:]) // clear the rest of the buffer
	}
	buf.readN += int64(n)
	return n, nil
}

// WriteTo does not enforce content length.
func (buf *Body) WriteTo(w io.Writer) (n int64, err error) {
	if buf.buf == nil {
		return 0, nil // nothing to write
	}
	return buf.buf.WriteTo(w)
}

func NewBody(buf *bufio.Reader, cl int64) *Body {
	return &Body{
		buf:           buf,
		contentLength: cl,
	}
}

package http

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
)

type Body struct {
	buf           *bufio.Reader
	readN         int64
	contentLength int64

	tmpFile      *os.File
	tmpCompleted bool  // whether the whole body has been written to the temporary file
	offset       int64 // offset in the temporary file OR offset in reading FROM the tmep file, use for writing the whole thing
}

func (buf *Body) Read(p []byte) (n int, err error) {
	if buf.tmpCompleted {
		if config.DefaultConfig.Debug {
			slog.Info("http body: reading from completed temporary file", "file", buf.tmpFile.Name())
		}
		n, err = buf.tmpFile.ReadAt(p, buf.offset)
		if err != nil && err != io.EOF {
			if config.DefaultConfig.Debug {
				slog.Error("http body: failed to read from temporary file", "err", err.Error())
			}
			return n, err
		}
		buf.offset += int64(n)
		if buf.offset >= buf.contentLength {
			if config.DefaultConfig.Debug {
				slog.Info("http body: completed reading from temporary file", "file", buf.tmpFile.Name())
			}
			buf.offset = 0 // reset offset for next read
			return n, nil
		}
	}

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

	if buf.tmpFile == nil {
		buf.tmpFile, err = os.CreateTemp("", "cap-http-request-body-*")
		if err != nil {
			if config.DefaultConfig.Debug {
				slog.Error("http body: failed to create temporary file", "err", err.Error())
			}
			return n, err
		}
		buf.offset = 0
	}
	_, err = buf.tmpFile.WriteAt(p[:n], buf.offset)
	if err != nil {
		if config.DefaultConfig.Debug {
			slog.Error("http body: failed to write to temporary file", "err", err.Error())
		}
		return n, err
	}
	buf.offset += int64(n)
	if buf.offset >= buf.contentLength {
		if config.DefaultConfig.Debug {
			slog.Info("http body: completed writing to temporary file", "file", buf.tmpFile.Name())
		}
		buf.tmpCompleted = true
		buf.offset = 0
	}

	return n, nil
}

// WriteTo does not enforce content length.
func (buf *Body) WriteTo(w io.Writer) (n int64, err error) {
	if buf.tmpCompleted {
		if config.DefaultConfig.Debug {
			slog.Info("http body: writing (writeto) from completed temporary file", "file", buf.tmpFile.Name())
		}
		return buf.tmpFile.WriteTo(w)
	}
	if buf.buf == nil {
		return 0, nil // nothing to write
	}
	return io.Copy(w, io.LimitReader(buf.buf, buf.contentLength))
}

func (b *Body) ContentLength() int64 {
	return b.contentLength
}

// this func is named close body in order to avoid ncruces/go-sqlite3 from closing it? (that's not even documented behavior :/)
func (buf *Body) CloseBody() error {
	fmt.Println("closing body")
	fmt.Println(runtime.Caller(1))
	if buf.tmpFile != nil {
		if config.DefaultConfig.Debug {
			slog.Info("http body: closing temporary file", "file", buf.tmpFile.Name())
		}
		return buf.tmpFile.Close()
	}
	return nil
}

func NewBody(buf *bufio.Reader, cl int64) *Body {
	return &Body{
		buf:           buf,
		contentLength: cl,
	}
}

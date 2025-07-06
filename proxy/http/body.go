package http

import (
	"bufio"
	"io"
	"log/slog"
	"os"

	"github.com/tiredkangaroo/cap/proxy/config"
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
		n, err = buf.tmpFile.ReadAt(p, buf.offset)
		if err != nil && err != io.EOF {
			if config.DefaultConfig.Debug {
				slog.Error("http body: failed to read from temporary file", "err", err.Error())
			}
			return n, err
		}
		buf.offset += int64(n)
		if buf.offset >= buf.contentLength {
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
		buf.tmpCompleted = true
		buf.offset = 0
	}

	return n, nil
}

func (buf *Body) WriteTo(w io.Writer) (n int64, err error) {
	if buf.tmpCompleted {
		return buf.tmpFile.WriteTo(w)
	}

	if buf.buf == nil {
		return 0, nil // nothing to write
	}

	if buf.tmpFile == nil {
		buf.tmpFile, err = os.CreateTemp("", "cap-http-request-body-*")
		if err != nil {
			if config.DefaultConfig.Debug {
				slog.Error("http body: failed to create temporary file", "err", err.Error())
			}
			return 0, err
		}
		buf.offset = 0
	}

	// Copy from buffer to tmp file until contentLength is reached
	tmpBuf := make([]byte, 32*1024)
	for buf.readN < buf.contentLength {
		maxRead := int64(len(tmpBuf))
		if buf.contentLength-buf.readN < maxRead {
			maxRead = buf.contentLength - buf.readN
		}
		readN, readErr := buf.buf.Read(tmpBuf[:maxRead])
		if readN > 0 {
			_, writeErr := buf.tmpFile.WriteAt(tmpBuf[:readN], buf.offset)
			if writeErr != nil {
				if config.DefaultConfig.Debug {
					slog.Error("http body: failed to write to temporary file", "err", writeErr.Error())
				}
				return n, writeErr
			}
			buf.offset += int64(readN)
			buf.readN += int64(readN)
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return n, readErr
		}
	}

	buf.tmpCompleted = true
	buf.offset = 0
	buf.buf = nil // release memory

	// Reset file read pointer and write to output
	_, err = buf.tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		if config.DefaultConfig.Debug {
			slog.Error("http body: failed to seek in temporary file", "err", err.Error())
		}
		return 0, err
	}
	return io.Copy(w, buf.tmpFile)
}

func (b *Body) ContentLength() int64 {
	return b.contentLength
}

// this func is named close body in order to avoid ncruces/go-sqlite3 from closing it? (that's not even documented behavior :/)
// NOTE: instead of renaming closebody use the new utils.go NoOpCloser
func (buf *Body) CloseBody() error {
	if buf.tmpFile != nil {
		return buf.tmpFile.Close()
	}
	buf.buf = nil // release the buffer
	return nil
}

func NewBody(buf *bufio.Reader, cl int64) *Body {
	return &Body{
		buf:           buf,
		contentLength: cl,
	}
}

package http

import (
	"bufio"
	"io"
	"log/slog"
	"os"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
)

type Body struct {
	buf           *bufio.Reader
	readN         int64
	contentLength int64
	tmpFile       *os.File
	tmpFilePath   string
	readDone      bool
}

// Read reads from the buffer, writing into a tempfile as it goes.
func (b *Body) Read(p []byte) (n int, err error) {
	if b.readN >= b.contentLength || b.buf == nil {
		b.buf = nil
		return 0, io.EOF
	}

	n, err = b.buf.Read(p)
	if err != nil && err != io.EOF {
		return n, err
	}

	maxRead := int(b.contentLength - b.readN)
	if n > maxRead {
		n = maxRead
		if config.DefaultConfig.Debug {
			slog.Warn("http body: read more than content length", "readN", b.readN, "contentLength", b.contentLength, "n", n)
		}
		clear(p[n:])
	}
	b.readN += int64(n)

	// Write to tempfile
	if !b.readDone {
		if b.tmpFile == nil {
			tmp, err := os.CreateTemp("", "bodytemp-*")
			if err != nil {
				return n, err
			}
			b.tmpFile = tmp
			b.tmpFilePath = tmp.Name()
		}
		_, _ = b.tmpFile.Write(p[:n])
		if b.readN >= b.contentLength {
			b.readDone = true
			b.tmpFile.Sync()
			b.tmpFile.Close()
		}
	}
	return n, nil
}

// WriteTo writes from the tempfile if available and fully read.
func (b *Body) WriteTo(w io.Writer) (n int64, err error) {
	// If we're done reading and have a tempfile, replay it
	if b.readDone && b.tmpFilePath != "" {
		file, err := os.Open(b.tmpFilePath)
		if err != nil {
			return 0, err
		}
		defer file.Close()
		return io.Copy(w, io.LimitReader(file, b.contentLength))
	}

	// Otherwise, copy current content and mirror into file
	if b.buf == nil {
		return 0, nil
	}

	tmp, err := os.CreateTemp("", "bodytemp-*")
	if err != nil {
		return 0, err
	}
	b.tmpFile = tmp
	b.tmpFilePath = tmp.Name()

	tee := io.TeeReader(io.LimitReader(b.buf, b.contentLength-b.readN), tmp)
	n64, err := io.Copy(w, tee)
	b.readN += n64
	b.readDone = true

	tmp.Sync()
	tmp.Close()

	return n64, err
}

func (b *Body) ContentLength() int64 {
	if b.contentLength < 0 {
		return 0
	}
	return b.contentLength
}

func (b *Body) Close() error {
	if b.tmpFilePath != "" {
		err := os.Remove(b.tmpFilePath)
		b.tmpFilePath = ""
		b.tmpFile = nil
		return err
	}
	return nil
}

func NewBody(buf *bufio.Reader, cl int64) *Body {
	return &Body{
		buf:           buf,
		contentLength: cl,
	}
}

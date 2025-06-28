package http

import (
	"io"
)

type Header map[string][]string

func (header Header) Get(k string) string {
	if values, ok := header[k]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

func (header Header) Set(k, v string) {
	if values, ok := header[k]; ok && len(values) > 0 {
		header[k][0] = v
	} else {
		header[k] = []string{v}
	}
}

func (header Header) Add(k, v string) {
	if values, ok := header[k]; ok {
		header[k] = append(values, v)
	} else {
		header[k] = []string{v}
	}
}

func (header Header) Del(k string) {
	// delete is a no-op if the key does not exist, so no need to check
	delete(header, k)
}

func (header Header) write(w io.Writer) error {
	// headers
	for key, values := range header {
		for _, value := range values {
			headerLine := make([]byte, 0, len(key)+len(value)+4)
			headerLine = append(headerLine, s2b(key)...)
			headerLine = append(headerLine, ": "...)
			headerLine = append(headerLine, s2b(value)...)
			headerLine = append(headerLine, '\r', '\n')
			_, err := w.Write(headerLine)
			if err != nil {
				return err
			}
		}
	}
	// end of headers
	_, err := w.Write([]byte{'\r', '\n'})
	return err
}

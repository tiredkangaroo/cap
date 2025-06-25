package http

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

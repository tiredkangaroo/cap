package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

// perform performs the request given, for HTTP and HTTPS connections. it can be used to add
// functionality later on.
func perform(config *Config, req *http.Request) ([]byte, error) {
	req.RequestURI = ""
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Proxy-Connection")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("req do error: %w", err)
	}

	// might be too resource heavy to do it this way
	data, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, fmt.Errorf("dump server response: %w", err)
	}
	return data, nil
}

package main

import (
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var URLS = []string{
	"https://example.com",
	"http://localhost:6200",
	"http://localhost:6200",
	"http://example.com",
	"https://localhost:6201",
	"http://google.com",
	"https://localhost:6201",
	"https://google.com",
	"https://localhost:6201",
	"https://apple.com",
	"https://yahoo.com",
	"https://example.org",
	"https://not-a-real-host.localhost:0000",
}

var PROXY_URL, _ = url.Parse("http://localhost:8000")

func main() {
	times, err := strconv.Atoi(os.Getenv("X_TIMES"))
	if err != nil {
		times = 10
		slog.Warn("X_TIMES not set or invalid, defaulting to 10", "error", err, "xtimes", os.Getenv("X_TIMES"))
	}

	os.Setenv("HTTP_PROXY", PROXY_URL.String())
	os.Setenv("HTTPS_PROXY", PROXY_URL.String())
	os.Setenv("NO_PROXY", "")

	transport := &http.Transport{
		Proxy: http.ProxyURL(PROXY_URL),
	}
	http.DefaultClient.Transport = transport

	for range times {
		for _, url := range URLS {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				slog.Error("error creating request", "error", err, "url", url)
				continue
			}
			if _, err := http.DefaultClient.Do(req); err != nil {
				slog.Error("do error", "url", url, "err", err)
			}
		}
	}
}

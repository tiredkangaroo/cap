package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type HackHandler struct{}

func (h HackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnavailableForLegalReasons)
	w.Write([]byte("you've been hacked?! mwhaha"))
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		data, err := json.Marshal(map[string]any{
			"proto":  r.Proto,
			"method": r.Method,
			"header": r.Header,
			"path":   r.URL.Path,
			"query":  r.URL.Query(),
			"body":   string(body), // haha thought there was an issue in my body in the proxy (gets encoded), but its the test server :/
		})
		if err != nil {
			http.Error(w, "error marshaling JSON", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			http.Error(w, "error writing response", http.StatusInternalServerError)
			return
		}
	})

	go func() {
		if err := http.ListenAndServe(":6200", nil); err != nil {
			log.Fatalf("http server error :6200 : %v", err)
		}
	}()
	if _, err := os.Stat("localhost.key"); err == nil {
		go func() {
			if err := http.ListenAndServeTLS(":6201", "localhost.crt", "localhost.key", nil); err != nil {
				log.Fatalf("https server error :6201 : %v", err)
			}
		}()
		go func() {
			if err := http.ListenAndServeTLS(":6202", "localhost.crt", "localhost.key", new(HackHandler)); err != nil {
				log.Fatalf("http server error :6202 HTTPS : %v", err)
			}
		}()
	} else {
		log.Default().Printf("stat failed for localhost.key: %s (:6201 unavailable, :6202 running http only)\n", err.Error())
		go func() {
			if err := http.ListenAndServe(":6202", new(HackHandler)); err != nil {
				log.Fatalf("http server error :6202 HTTP : %v", err)
			}
		}()
	}
	select {} // block forever
}

package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		data, err := json.Marshal(map[string]any{
			"proto":  r.Proto,
			"method": r.Method,
			"header": r.Header,
			"path":   r.URL.Path,
			"query":  r.URL.Query(),
			"body":   body,
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
			log.Fatalf("http server error: %v", err)
		}
	}()
	go func() {
		if err := http.ListenAndServeTLS(":6201", "localhost.crt", "localhost.key", nil); err != nil {
			log.Fatalf("https server error: %v", err)
		}
	}()
	select {} // block forever
}

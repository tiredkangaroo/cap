package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func startControlServer(config *Config) {
	// Start the control server
	http.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(config)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("failed to marshal config"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	http.HandleFunc("POST /config", func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("failed to read request body"))
			return
		}

		// possible validation of the config here and authority to change it

		var newConfig Config
		err = json.Unmarshal(b, &newConfig)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("failed to decode config"))
			return
		}

		*config = newConfig
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("config updated"))
	})

	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		panic(err)
	}
}

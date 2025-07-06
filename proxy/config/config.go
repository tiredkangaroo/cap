package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var DefaultConfig = &Config{}

const proxy_config_file = "PROXY_CONFIG_FILE"

// Config is the configuration for the proxy.
type Config struct {
	// Debug is a boolean that determines whether the proxy should run in debug mode.
	Debug bool `json:"debug"`
	// RealIPHeader is a boolean that determines whether the proxy should add the IP
	// address of the client in the X-Forwarded-For header.
	RealIPHeader bool `json:"real_ip_header"`
	// CertificateLifetime is the lifetime of the certificate in hours. It is possible
	// to set it to less than 0 in which any new certificates generated will NEVER be valid.
	CertificateLifetime int `json:"certificate_lifetime"`
	// MITM determines who is responsible for the TLS connection. If true, the responsibility
	// is on the proxy. If false, the responsibility is on the client.
	//
	// MITM is useful for inspecting the traffic between the client and the server. However,
	// it is more resource intensive to generate and store the certificates for each host, perform a
	// TLS handshake as well as to decrypt the traffic, reencrypt it, move requests and responses.
	MITM bool `json:"mitm"`
	// PerformDelay is the delay in milliseconds before the proxy performs the request. It must be a positive
	// integer. This can be useful for testing purposes, such as simulating network latency, slowing down
	// actions performed, or other debugging purposes.
	PerformDelay uint `json:"perform_delay"`

	// RequireApproval is a boolean that determines whether the proxy should require approval to perform
	// requests. If true, the proxy will wait for an approval before performing the request. The client
	// may timeout if the approval is not received in time. This should be taken into consideration.
	//
	// This is useful for debugging purposes, such as inspecting the request before it is performed,
	// or for security purposes, such as ensuring that the request is safe to perform.
	RequireApproval bool `json:"require_approval"`

	ProvideRequestBody  bool `json:"provide_request_body"`
	ProvideResponseBody bool `json:"provide_response_body"`

	// GetClientProcessInfo is a boolean that determines whether the proxy should provide information about the
	// client process that made the request. This includes the process ID and the name of the application.
	//
	// This can be useful for both debugging and production purposes, however this operation take a significant
	// amount of time to perform, since it requires the proxy to query the operating system by running shell
	// commands for the process information.
	GetClientProcessInfo bool `json:"get_client_process_info"`

	// TimelineBasedStateUpdates is a boolean that determines whether the proxy should send state updates to the client
	// based on timeline events. If true, the proxy will send updates to the client whenever a major or minor timeline event
	// occurs.
	TimelineBasedStateUpdates bool `json:"timeline_based_state_updates"`
}

func init() {
	DefaultConfig.Debug = os.Getenv("DEBUG") == "true"

	var dirname string
	if os.Getenv("BUILT") != "false" {
		execPath, err := os.Executable()
		if err != nil {
			slog.Error("failed to get executable path (nonfatal)", "err", err.Error())
		} else {
			dirname = filepath.Dir(execPath)
		}
	} else {
		dirname = "."
	}

	filename := os.Getenv(proxy_config_file)
	if filename == "" {
		slog.Warn("no proxy config file specified, using config.json as default")
		filename = fmt.Sprintf("%s/config.json", dirname)
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		slog.Error("creating/opening config file specified", "specified_file", proxy_config_file, "err", err.Error())
		return
	}

	if err := setConfigFromFile(file, DefaultConfig); err != nil {
		slog.Error("setting config from file specified", "specified_file", proxy_config_file, "err", err.Error())
	}

	saveConfigFile(file, DefaultConfig)
}

func setConfigFromFile(file *os.File, config *Config) error {
	rf, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if len(rf) == 0 {
		slog.Info("setting default config")
		config.CertificateLifetime = 7200
		config.GetClientProcessInfo = true
		config.ProvideRequestBody = true
		config.ProvideResponseBody = true
		config.RealIPHeader = true
		return nil // empty config file, nothing to do
	}

	err = json.Unmarshal(rf, config)
	if err != nil {
		return err
	}
	return nil
}

func saveConfigFile(file *os.File, config *Config) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		defer os.Exit(0)
		<-c
		slog.Info("received signal, saving config file")
		data, err := json.Marshal(config)
		if err != nil {
			slog.Error("saving config file at marshal step", "err", err.Error())
			return
		}
		if err := file.Truncate(0); err != nil {
			slog.Error("saving config file at truncate step", "err", err.Error())
		}
		if _, err := file.Seek(0, 0); err != nil {
			slog.Error("saving config file at the seek 0 step", "err", err.Error())
		}
		if _, err := file.Write(data); err != nil {
			slog.Error("saving config file at file.Write step", "err", err.Error())
			return
		}
	}()
}

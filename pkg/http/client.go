package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// ClientConfig holds configuration settings for the HTTP client.
type ClientConfig struct {
	// Timeout specifies the timeout for reading HTTP headers in seconds.
	Timeout int `mapstructure:"timeout"`
}

// NewClient creates a new http client based on the provided configuration.
func NewClient(cfg ClientConfig) (*http.Client, error) {
	slog.Info("Initializing HTTP client for the service", "timeout", cfg.Timeout)

	// Validate the configuration
	if cfg.Timeout <= 0 {
		return nil, fmt.Errorf("http client timeout is invalid: %d", cfg.Timeout)
	}

	return &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}, nil
}

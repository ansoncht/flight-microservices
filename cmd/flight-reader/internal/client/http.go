package client

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/config"
)

func NewHTTP() (*http.Client, error) {
	slog.Info("Creating HTTP client for the service")

	cfg, err := config.LoadHTTPClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load HTTP client config: %w", err)
	}

	return &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}, nil
}

package client

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/config"
)

func NewHTTP(cfg config.HTTPClientConfig) (*http.Client, error) {
	slog.Info("Creating HTTP client for the service")

	return &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}, nil
}

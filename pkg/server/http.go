package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// HTTPConfig holds configuration settings for the HTTP server.
type HTTPConfig struct {
	// Port specifies the port where the HTTP server listens for requests.
	Port string `mapstructure:"port"`
	// Timeout specifies the timeout for reading HTTP headers in seconds.
	Timeout int `mapstructure:"timeout"`
}

// HTTP holds the HTTP server instance and its dependencies.
type HTTP struct {
	server *http.Server
}

// NewHTTPServer creates a new HTTP server instance based on the provided configuration.
func NewHTTPServer(cfg HTTPConfig, handler http.Handler) (*HTTP, error) {
	slog.Info("Initializing HTTP server for the service", "port", cfg.Port, "timeout", cfg.Timeout)

	if handler == nil {
		return nil, fmt.Errorf("handler is nil")
	}

	// Validate the configuration
	if cfg.Timeout <= 0 {
		return nil, fmt.Errorf("http server timeout is invalid: %d", cfg.Timeout)
	}

	if cfg.Port == "" {
		return nil, fmt.Errorf("port number is empty")
	}

	port, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("port number is invalid: %w", err)
	}
	if port < 1 {
		return nil, fmt.Errorf("port number must be greater than 0")
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(cfg.Timeout) * time.Second,
	}

	return &HTTP{
		server: server,
	}, nil
}

// Serve starts the HTTP server and handles incoming requests.
func (h *HTTP) Serve(ctx context.Context) error {
	slog.Info("Starting HTTP server", "port", h.server.Addr)

	c := make(chan error)

	// Start the server in a goroutine
	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start HTTP server", "error", err)
			c <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Stopping HTTP server due to context cancellation")
		return fmt.Errorf("context canceled while running HTTP server: %w", ctx.Err())
	case err := <-c:
		slog.Error("Failed to start HTTP server", "error", err)
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
}

// Close gracefully shuts down the HTTP server.
func (h *HTTP) Close(ctx context.Context) error {
	if err := h.server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown HTTP server", "error", err)
		return fmt.Errorf("failed to shutdown: %w", err)
	}

	return nil
}

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/client"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/fetcher"
)

// HTTP struct represents the HTTP server and its dependencies.
type HTTP struct {
	server     *http.Server       // HTTP server for incoming request
	grpcClient *client.GrpcClient // gRPC client for communication
	fetchers   []fetcher.Fetcher  // List of fetchers for processing flights
}

// NewHTTP creates a new HTTP server instance.
func NewHTTP(
	cfg config.HTTPServerConfig,
	grpcClient *client.GrpcClient,
	fetchers []fetcher.Fetcher,
) (*HTTP, error) {
	slog.Info("Creating HTTP server for the service")

	h := &HTTP{
		grpcClient: grpcClient,
		fetchers:   fetchers,
	}

	// Register endpoints for the server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/fetch", h.handler())

	// Validate the port number
	if cfg.Port == "" {
		return nil, fmt.Errorf("empty port number")
	}

	port, _ := strconv.Atoi(cfg.Port)
	if port < 1 {
		return nil, fmt.Errorf("port number must be greater than 1")
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(cfg.Timeout) * time.Second,
	}

	h.server = server

	return h, nil
}

// ServeHTTP starts the HTTP server and handles incoming requests.
func (h *HTTP) ServeHTTP(ctx context.Context) error {
	slog.Info("Starting HTTP server", "port", h.server.Addr)

	c := make(chan error)

	// Start the server in a goroutine
	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start HTTP server", "error", err)
			c <- fmt.Errorf("failed to start HTTP server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Stopping HTTP server due to context cancellation")
		return nil
	case err := <-c:
		return err
	}
}

// Close gracefully shuts down the HTTP server.
func (h *HTTP) Close(ctx context.Context) error {
	if err := h.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown: %w", err)
	}

	return nil
}

// handler returns an HTTP handler function for fetching flights.
func (h *HTTP) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Endpoint initiates manual flights fetching")

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Process flights using the fetchers
		if err := fetcher.ProcessFlights(r.Context(), h.grpcClient, h.fetchers, "VHHH"); err != nil {
			slog.Error("Failed to process flight", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Successfully triggered a manual fetch")
	}
}

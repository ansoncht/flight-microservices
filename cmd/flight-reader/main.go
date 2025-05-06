package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansoncht/flight-microservices/internal/reader/client"
	"github.com/ansoncht/flight-microservices/internal/reader/config"
	"github.com/ansoncht/flight-microservices/internal/reader/service"

	appHTTP "github.com/ansoncht/flight-microservices/pkg/http"
	"github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/ansoncht/flight-microservices/pkg/logger"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Create a context that listens for OS interrupt signals (e.g., Ctrl+C)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		return
	}

	// Create a customized logger
	logger, err := logger.NewLogger(cfg.LoggerConfig)
	if err != nil {
		slog.Warn("Failed to create custom logger, using default logger instead", "error", err)
	}

	slog.SetDefault(&logger)

	httpClient, err := appHTTP.NewClient(cfg.HTTPClientConfig)
	if err != nil {
		slog.Error("Failed to create HTTP client", "error", err)
		return
	}

	// Create reader service to fetch flight and route data
	reader, err := initializeReaderService(
		cfg.FlightAPIClientConfig,
		cfg.RouteAPIClientConfig,
		cfg.KafkaWriterConfig,
		httpClient,
	)
	if err != nil {
		slog.Error("Failed to initialize reader service", "error", err)
		return
	}

	// Create a new HTTP server and handler
	httpServer, err := initializeHTTPServerWithHandler(cfg.HTTPServerConfig, reader)
	if err != nil {
		slog.Error("Failed to create HTTP server with handler", "error", err)
		return
	}

	// Run the server in background
	if err := startBackgroundJobs(ctx, httpServer); err != nil {
		slog.Error("Failed to start background jobs", "error", err)
		return
	}

	// Perform a safe shutdown
	if err := safeShutDown(ctx, httpClient, httpServer, reader); err != nil {
		slog.Error("Failed to perform graceful shutdown", "error", err)
		return
	}

	slog.Info("Flight Reader service has fully stopped")
}

// initializeHTTPServerWithHandler initializes the http server with a handler to trigger reader's workflow.
func initializeHTTPServerWithHandler(
	httpCfg appHTTP.ServerConfig,
	readerService *service.Reader,
) (*appHTTP.HTTP, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/fetch", readerService.HTTPHandler)

	httpServer, err := appHTTP.NewServer(httpCfg, mux)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP server: %w", err)
	}

	return httpServer, nil
}

// initializeReaderService initializes the reader service.
func initializeReaderService(
	flightCfg config.FlightAPIConfig,
	routeCfg config.RouteAPIConfig,
	kafkaCfg kafka.WriterConfig,
	httpClient *http.Client,
) (*service.Reader, error) {
	flightClient, err := client.NewFlightAPI(flightCfg, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create flight api client: %w", err)
	}

	routeClient, err := client.NewRouteAPI(routeCfg, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create route api client: %w", err)
	}

	kafkaWriter, err := kafka.NewKafkaWriter(kafkaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka writer: %w", err)
	}

	reader, err := service.NewReader(flightClient, routeClient, kafkaWriter)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader service: %w", err)
	}

	return reader, nil
}

// startBackgroundJobs starts the HTTP server in background.
func startBackgroundJobs(ctx context.Context, httpServer *appHTTP.HTTP) error {
	// Use errgroup to manage concurrent tasks
	g, gCtx := errgroup.WithContext(ctx)

	// Start the HTTP server
	g.Go(func() error {
		return httpServer.Serve(gCtx)
	})

	// Wait for the context to be done (e.g., due to an interrupt signal)
	<-gCtx.Done()

	// Wait for all goroutines to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to start one of the background jobs: %w", err)
	}

	return nil
}

// safeShutDown shuts down http client, http server and reader gracefully.
func safeShutDown(
	ctx context.Context,
	httpClient *http.Client,
	httpServer *appHTTP.HTTP,
	reader *service.Reader,
) error {
	// Attempt to close the HTTP server
	if err := httpServer.Close(ctx); err != nil {
		slog.Error("Failed to shutdown HTTP server", "error", err)
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	// Close the HTTP client
	httpClient.CloseIdleConnections()

	if err := reader.Close(); err != nil {
		slog.Error("Failed to shutdown reader", "error", err)
		return fmt.Errorf("failed to shutdown reader: %w", err)
	}

	// Return nil if all shutdowns were successful
	return nil
}

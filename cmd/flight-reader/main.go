package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ansoncht/flight-microservices/internal/reader/client"
	"github.com/ansoncht/flight-microservices/internal/reader/config"
	"github.com/ansoncht/flight-microservices/internal/reader/service"

	appHTTP "github.com/ansoncht/flight-microservices/pkg/http"
	"github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/ansoncht/flight-microservices/pkg/logger"
	"golang.org/x/sync/errgroup"
)

const timeout = 10 * time.Second

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

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.Info("Starting background jobs")
		return startBackgroundJobs(gCtx, httpServer)
	})

	g.Go(func() error {
		<-gCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		slog.Info("Shutting down background jobs")
		return safeShutDown(shutdownCtx, httpClient, httpServer, reader)
	})

	if err := g.Wait(); err != nil {
		slog.Error("Service exited with error", "error", err)
	}
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
	if err := httpServer.Serve(ctx); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
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

	httpClient.CloseIdleConnections()
	reader.Close()

	return nil
}

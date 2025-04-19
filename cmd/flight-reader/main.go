package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/client"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/fetcher"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/scheduler"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/server"
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

	// Create gRPC and HTTP clients
	grpcClient, httpClient, err := initializeClients(cfg.GrpcClientConfig, cfg.HTTPClientConfig)
	if err != nil {
		slog.Error("Failed to create clients", "error", err)
		return
	}

	// Create fetchers for scheduler and http server
	fetchers, err := initializeFetchers(cfg.FlightFetcherConfig, cfg.RouteFetcherConfig, httpClient)
	if err != nil {
		slog.Error("Failed to create fetchers", "error", err)
		return
	}

	// Create a HTTP server with the grpc client and fetchers
	httpServer, err := server.NewHTTP(cfg.HTTPServerConfig, grpcClient, fetchers)
	if err != nil {
		slog.Error("Failed to create HTTP server", "error", err)
		return
	}

	// Create a scheduler with the grpc client and fetchers
	scheduler, err := scheduler.NewScheduler(cfg.SchedulerConfig, grpcClient, fetchers)
	if err != nil {
		slog.Error("Failed to create scheduler", "error", err)
		return
	}

	// Run the server and schedule jobs concurrently
	if err := startBackgroundJobs(ctx, httpServer, scheduler); err != nil {
		slog.Error("Failed to run background jobs concurrently", "error", err)
		return
	}

	// Perform a safe shutdown of the server and clients
	if err := safeShutDown(ctx, grpcClient, httpClient, httpServer); err != nil {
		slog.Error("Failed to perform graceful shutdown", "error", err)
		return
	}

	slog.Info("Flight Reader service has fully stopped")
}

func initializeClients(
	grpcCfg config.GrpcClientConfig,
	httpCfg config.HTTPClientConfig,
) (*client.GrpcClient, *http.Client, error) {
	grpcClient, err := client.NewGRPC(grpcCfg)
	if err != nil {
		slog.Error("Failed to create gRPC client", "error", err)
		return nil, nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	httpClient, err := client.NewHTTP(httpCfg)
	if err != nil {
		slog.Error("Failed to create HTTP client", "error", err)
		return nil, nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return grpcClient, httpClient, nil
}

func initializeFetchers(
	flightCfg config.FlightFetcherConfig,
	routeCfg config.RouteFetcherConfig,
	httpClient *http.Client,
) ([]fetcher.Fetcher, error) {
	flightFetcher, err := fetcher.NewFlightFetcher(flightCfg, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create flight fetcher: %w", err)
	}

	routeFetcher, err := fetcher.NewRouteFetcher(routeCfg, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create route fetcher: %w", err)
	}

	return []fetcher.Fetcher{flightFetcher, routeFetcher}, nil
}

// startBackgroundJobs starts the HTTP server and scheduler concurrently.
func startBackgroundJobs(ctx context.Context, httpServer *server.HTTP, scheduler *scheduler.Scheduler) error {
	// Use errgroup to manage concurrent tasks
	g, gCtx := errgroup.WithContext(ctx)

	// Start the HTTP server and job scheduler concurrently
	g.Go(func() error {
		return httpServer.ServeHTTP(gCtx)
	})

	g.Go(func() error {
		return scheduler.ScheduleJob(gCtx)
	})

	// Wait for the context to be done (e.g., due to an interrupt signal)
	<-gCtx.Done()

	// Wait for all goroutines to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to run concurrent tasks: %w", err)
	}

	return nil
}

// safeShutDown shuts down clients and server gracefully.
func safeShutDown(
	ctx context.Context,
	grpcClient *client.GrpcClient,
	httpClient *http.Client,
	httpServer *server.HTTP,
) error {
	// Attempt to close the HTTP server
	if err := httpServer.Close(ctx); err != nil {
		slog.Error("Failed to shutdown HTTP server", "error", err)
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	// Close the HTTP client
	httpClient.CloseIdleConnections()

	// Attempt to close the gRPC client
	if err := grpcClient.Close(); err != nil {
		slog.Error("Failed to shutdown gRPC client", "error", err)
		return fmt.Errorf("failed to shutdown gRPC client: %w", err)
	}

	// Return nil if all shutdowns were successful
	return nil
}

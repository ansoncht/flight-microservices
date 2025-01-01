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
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/fetcher"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/scheduler"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/server"
	logger "github.com/ansoncht/flight-microservices/pkg/log"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Create a context that listens for OS interrupt signals (e.g., Ctrl+C)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize the application-wide logger
	err := logger.NewLogger()
	if err != nil {
		slog.Error("Failed to initialize custom logger", "error", err)
		return
	}

	// Create gRPC and HTTP clients
	grpcClient, err := client.NewGRPC()
	if err != nil {
		slog.Error("Failed to create gRPC client", "error", err)
		return
	}

	httpClient, err := client.NewHTTP()
	if err != nil {
		slog.Error("Failed to create HTTP client", "error", err)
		return
	}

	// Create fetchers for scheduler and http server
	fetchers, err := initializeFetchers(httpClient)
	if err != nil {
		slog.Error("Failed to create fetchers", "error", err)
		return
	}

	// Create a HTTP server with the grpc client and fetchers
	httpServer, err := server.NewHTTP(grpcClient, fetchers)
	if err != nil {
		slog.Error("Failed to create HTTP server", "error", err)
		return
	}

	// Create a scheduler with the grpc client and fetchers
	scheduler, err := scheduler.NewScheduler(grpcClient, fetchers)
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

func initializeFetchers(httpClient *http.Client) ([]fetcher.Fetcher, error) {
	flightFetcher, err := fetcher.NewFlightFetcher(httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create flight fetcher: %w", err)
	}

	routeFetcher, err := fetcher.NewRouteFetcher(httpClient)
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

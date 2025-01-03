package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/client"
	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/poster"
	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/server"
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

	// Create HTTP clients
	httpClient, err := client.NewHTTP()
	if err != nil {
		slog.Error("Failed to create HTTP client", "error", err)
		return
	}

	// Create posters for different social media
	threads, err := poster.NewThreadsClient(ctx, httpClient)
	if err != nil {
		slog.Error("Failed to create Threads poster", "error", err)
		return
	}

	twitter, err := poster.NewTwitterClient()
	if err != nil {
		slog.Error("Failed to create Twitter poster", "error", err)
		return
	}

	posters := []poster.Poster{threads, twitter}

	// Create a gRPC server
	grpcServer, err := server.NewGRPC(posters)
	if err != nil {
		slog.Error("Failed to create gRPC server", "error", err)
		return
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return grpcServer.ServeGRPC(gCtx)
	})

	<-gCtx.Done()

	if err := g.Wait(); err != nil {
		slog.Error("Encounter unexpected error", "error", err)
		return
	}

	grpcServer.Close()

	slog.Info("Flight Poster service has fully stopped")
}

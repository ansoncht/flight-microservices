package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/client"
	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/poster"
	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/server"
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

	// Create HTTP clients
	httpClient, err := client.NewHTTP(cfg.HTTPClientConfig)
	if err != nil {
		slog.Error("Failed to create HTTP client", "error", err)
		return
	}

	// Create posters for different social media
	posters, err := initializePosters(ctx, cfg.ThreadsClientConfig, cfg.TwitterClientConfig, httpClient)
	if err != nil {
		slog.Error("Failed to create posters", "error", err)
		return
	}

	// Create a gRPC server
	grpcServer, err := server.NewGRPC(cfg.GrpcServerConfig, posters)
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

func initializePosters(
	ctx context.Context,
	threadsCfg config.ThreadsClientConfig,
	twittercfg config.TwitterClientConfig,
	httpClient *http.Client,
) ([]poster.Poster, error) {
	// Create posters for different social media
	threads, err := poster.NewThreadsClient(ctx, threadsCfg, httpClient)
	if err != nil {
		slog.Error("Failed to create Threads poster", "error", err)
		return nil, fmt.Errorf("failed to create Threads poster: %w", err)
	}

	twitter, err := poster.NewTwitterClient(twittercfg)
	if err != nil {
		slog.Error("Failed to create Twitter poster", "error", err)
		return nil, fmt.Errorf("failed to create Twitter poster: %w", err)
	}

	return []poster.Poster{threads, twitter}, nil
}

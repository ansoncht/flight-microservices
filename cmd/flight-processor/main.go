package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/client"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/db"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/server"
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
		slog.Warn(
			"Failed to create custom logger, using default logger instead",
			"error", err,
		)
	}

	slog.SetDefault(&logger)

	// Create a Mongo client
	mongoDB, err := db.NewMongo(cfg.MongoClientConfig)
	if err != nil {
		slog.Error("Failed to create Mongo client", "error", err)
		return
	}

	// Create a gRPC client
	grpcClient, err := client.NewGRPC(cfg.GrpcClientConfig)
	if err != nil {
		slog.Error("Failed to create gRPC client", "error", err)
		return
	}

	// Create a gRPC server
	grpcServer, err := server.NewGRPC(cfg.GrpcServerConfig, mongoDB, grpcClient)
	if err != nil {
		slog.Error("Failed to create gRPC server", "error", err)
		return
	}

	// Run the server and establish mongo connection concurrently
	if err := startBackgroundJobs(ctx, grpcServer, mongoDB); err != nil {
		slog.Error("Failed to run background jobs concurrently", "error", err)
		return
	}

	// Perform a safe shutdown of the server and clients
	if err := safeShutDown(ctx, grpcServer, mongoDB); err != nil {
		slog.Error("Failed to perform graceful shutdown", "error", err)
		return
	}

	slog.Info("Flight Processor service has fully stopped")
}

// startBackgroundJobs starts the gRPC server and Mongo client concurrently.
func startBackgroundJobs(ctx context.Context, grpcServer *server.GrpcServer, mongoDB *db.Mongo) error {
	// Use errgroup to manage concurrent tasks
	g, gCtx := errgroup.WithContext(ctx)

	// Start the gRPC server and Mongo client concurrently
	g.Go(func() error {
		return grpcServer.ServeGRPC(gCtx)
	})

	g.Go(func() error {
		return mongoDB.Connect(gCtx)
	})

	// Wait for the context to be done (e.g., due to an interrupt signal)
	<-gCtx.Done()

	// Wait for all goroutines to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to run concurrent tasks: %w", err)
	}

	return nil
}

// safeShutDown shut down clients and server gracefully.
func safeShutDown(ctx context.Context, grpcServer *server.GrpcServer, mongoDB *db.Mongo) error {
	slog.Info("Shutting down gRPC server and MongoDB client")

	grpcServer.Close()

	if err := mongoDB.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to shutdown mongodb client: %w", err)
	}

	return nil
}

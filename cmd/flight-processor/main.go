package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/db"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/server"
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

	// Create a Mongo client
	mongoDB, err := db.NewMongo()
	if err != nil {
		slog.Error("Failed to create Mongo client", "error", err)
		return
	}

	// Create a gRPC server
	grpcServer, err := server.NewGRPC(mongoDB)
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
	grpcServer.Close()

	if err := mongoDB.Disconnect(ctx); err != nil {
		slog.Error("Failed to shutdown Mongo client", "error", err)
		return fmt.Errorf("failed to shutdown mongo client: %w", err)
	}

	return nil
}

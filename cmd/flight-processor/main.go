package main

import (
	"context"
	"fmt"
	"log"
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
	// create context to listen to os signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// make app-wide logger
	err := logger.MakeLogger()
	if err != nil {
		log.Panicln(err)
	}

	mongoDB, err := db.NewMongoDB()
	if err != nil {
		log.Panicln(err)
	}

	grpcServer, err := server.NewGRPCServer(mongoDB)
	if err != nil {
		log.Panicln(err)
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return grpcServer.ServeGRPC(gCtx)
	})

	g.Go(func() error {
		return mongoDB.Connect(gCtx)
	})

	<-gCtx.Done()

	if err := safeShutDown(ctx, grpcServer, mongoDB); err != nil {
		log.Panicln(err)
	}

	if err := g.Wait(); err != nil {
		slog.Error("Encounter unexpected error", "error", err)

		log.Panicln(err)
	}

	slog.Info("flight processor has fully stopped")
}

// safeShutDown turns off clients and server gracefully.
func safeShutDown(ctx context.Context, grpcServer *server.GRPCServer, mongoDB *db.MongoDB) error {
	grpcServer.Close()

	if err := mongoDB.Disconnect(ctx); err != nil {
		slog.Error("Failed to shutdown MongoDB connection", "error", err)

		return fmt.Errorf("failed to disconnect from mongo db: %w", err)
	}

	return nil
}

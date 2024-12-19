package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansoncht/flight-microservices/cmd/flight-poster/server"
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
		slog.Error("Failed to create custom logger", "error", err)
		return
	}

	grpcServer, err := server.NewGRPCServer()
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

package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/server"
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

	grpcServer, err := server.NewGRPCServer()
	if err != nil {
		log.Panicln(err)
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return grpcServer.ServeGRPC()
	})

	<-gCtx.Done()

	if err := g.Wait(); err != nil {
		slog.Error("Encounter unexpected error", "error", err)
		log.Panicln(err)
	}

	slog.Info("flight processor has fully stopped")
}

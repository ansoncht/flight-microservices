package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/clients"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/server"
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

	grpcClient, err := clients.NewGRPCClient()
	if err != nil {
		log.Panicln(err)
	}

	httpClient, err := clients.NewHTTPClient()
	if err != nil {
		log.Panicln(err)
	}

	httpServer, err := server.NewHTTPServer(httpClient, grpcClient)
	if err != nil {
		log.Panicln(err)
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return httpServer.ServeHTTP()
	})

	g.Go(func() error {
		return httpServer.ScheduleJob(gCtx)
	})

	<-gCtx.Done()

	if err := g.Wait(); err != nil {
		slog.Error("Encounter unexpected error", "error", err)
		log.Panicln(err)
	}

	slog.Info("flight reader has fully stopped")
}

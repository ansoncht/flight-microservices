package main

import (
	"context"
	"fmt"
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
		return httpServer.ServeHTTP(gCtx)
	})

	g.Go(func() error {
		return httpServer.ScheduleJob(gCtx)
	})

	<-gCtx.Done()

	if err := safeShutDown(ctx, grpcClient, httpClient, httpServer); err != nil {
		slog.Error("Failed to perform graceful shutdown", "error", err)
		log.Panicln(err)
	}

	if err := g.Wait(); err != nil {
		slog.Error("Encounter unexpected error", "error", err)
		log.Panicln(err)
	}

	slog.Info("flight reader has fully stopped")
}

// safeShutDown turns off clients and server gracefully.
func safeShutDown(
	ctx context.Context,
	grpcClient *clients.GRPCClient,
	httpClient *clients.HTTPClient,
	httpServer *server.Server,
) error {
	// close the http server
	if err := httpServer.Close(ctx); err != nil {
		slog.Error("Failed to shutdown HTTP server", "error", err)

		return fmt.Errorf("failed to gracefully shutdown http server")
	}

	// close the http client
	httpClient.Close()

	// close the gRPC client
	if err := grpcClient.Close(); err != nil {
		slog.Error("Failed to shutdown gRPC client", "error", err)

		return fmt.Errorf("failed to gracefully shutdown gRPC client")
	}

	return nil
}

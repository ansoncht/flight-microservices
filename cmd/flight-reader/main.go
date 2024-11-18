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
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/server"
	"golang.org/x/sync/errgroup"
)

func main() {
	// create context to listen to os signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// make app-wide logger
	err := makeLogger()
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

// makeLogger create and instantiate a default logger.
func makeLogger() error {
	slog.Debug("Creating logger for the service")

	loggerCfg, err := config.MakeLoggerConfig()
	if err != nil {
		return fmt.Errorf("failed to get logger config: %w", err)
	}

	// set slog level
	var level slog.Leveler
	addSource := false
	switch loggerCfg.Level {
	case "debug":
		level = slog.LevelDebug
		addSource = true
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return fmt.Errorf("invalid log level: %s", loggerCfg.Level)
	}

	// set slog source logging
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	}

	// set slog format
	handler = slog.NewTextHandler(os.Stdout, opts)
	if loggerCfg.JSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))

	return nil
}

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/client"
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

	httpClient, err := makeHTTPClient()
	if err != nil {
		log.Panicln(err)
	}

	httpServer, err := makeHTTPServer(httpClient)
	if err != nil {
		log.Panicln(err)
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return server.ServeHTTP(httpServer)
	})

	if err := httpClient.FetchFlightsFromAPI(gCtx); err != nil {
		log.Panicln(err)
	}

	<-gCtx.Done()

	if err := g.Wait(); err != nil {
		slog.Error("unexpected error occur", "error", err)
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

// makeHTTPServer create and instantiate a http server.
func makeHTTPServer(httpClient *client.HTTPClient) (*http.Server, error) {
	slog.Debug("Creating http server for the service")

	httpCfg, err := config.MakeHTTPServerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get http server config: %w", err)
	}

	if port := httpCfg.Port; port == "" {
		return nil, fmt.Errorf("empty port number")
	}

	// register endpoints.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/fetch", server.FetchHandler(httpClient))

	return &http.Server{
		Addr:              ":" + httpCfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(httpCfg.ReadHeaderTimeout) * time.Second,
	}, nil
}

// makeHTTPClient create and instantiate a http client.
func makeHTTPClient() (*client.HTTPClient, error) {
	slog.Debug("Creating http client for the service")

	httpCfg, err := config.MakeHTTPClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get http client config: %w", err)
	}

	return &client.HTTPClient{
		Client: &http.Client{
			Timeout: time.Duration(httpCfg.Timeout) * time.Second,
		},
		Endpoint: httpCfg.URL,
	}, nil
}

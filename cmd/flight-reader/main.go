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
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/scheduler"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/server"
	pb "github.com/ansoncht/flight-microservices/proto/src/summarizer"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	grpcClient, err := makeGRPCClient()
	if err != nil {
		log.Panicln(err)
	}

	httpClient, err := makeHTTPClient(grpcClient)
	if err != nil {
		log.Panicln(err)
	}

	httpServer, err := makeHTTPServer(ctx, httpClient)
	if err != nil {
		log.Panicln(err)
	}

	schedule := makeScheduler(httpClient)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return server.ServeHTTP(httpServer)
	})

	g.Go(func() error {
		return schedule.ScheduleDailyFetch(gCtx)
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

// makeHTTPServer create and instantiate a http server.
func makeHTTPServer(ctx context.Context, httpClient *client.HTTPClient) (*http.Server, error) {
	slog.Debug("Creating http server for the service")

	cfg, err := config.MakeHTTPServerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get http server config: %w", err)
	}

	if port := cfg.Port; port == "" {
		return nil, fmt.Errorf("empty port number")
	}

	// register endpoints.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/fetch", server.FetchHandler(ctx, httpClient))

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
	}, nil
}

// makeHTTPClient create and instantiate a http client.
func makeHTTPClient(grpc pb.SummarizerClient) (*client.HTTPClient, error) {
	slog.Debug("Creating http client for the service")

	cfg, err := config.MakeHTTPClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get http client config: %w", err)
	}

	return &client.HTTPClient{
		Client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		Endpoint: cfg.URL,
		GRPC:     grpc,
	}, nil
}

// makeGRPCClient create and instantiate a gRPC client.
func makeGRPCClient() (pb.SummarizerClient, error) {
	slog.Debug("Creating gRPC client for the service")

	cfg, err := config.MakeGRPCClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get http gRPC config: %w", err)
	}

	conn, err := grpc.NewClient(cfg.URL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return pb.NewSummarizerClient(
		conn,
	), nil
}

// makeScheduler create and instantiate a scheduler.
func makeScheduler(httpClient *client.HTTPClient) *scheduler.Scheduler {
	slog.Debug("Creating scheduler for the service")

	return &scheduler.Scheduler{
		FlightFetcher: httpClient,
	}
}

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansoncht/flight-microservices/internal/poster/client"
	"github.com/ansoncht/flight-microservices/internal/poster/service"
	appHTTP "github.com/ansoncht/flight-microservices/pkg/http"
	"github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/ansoncht/flight-microservices/pkg/mongo"
	"github.com/ansoncht/flight-microservices/pkg/repository"

	"github.com/ansoncht/flight-microservices/internal/poster/config"
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
		slog.Warn("Failed to create custom logger, using default logger instead", "error", err)
	}

	slog.SetDefault(&logger)

	httpClient, err := appHTTP.NewClient(cfg.HTTPClientConfig)
	if err != nil {
		slog.Error("Failed to create HTTP client", "error", err)
		return
	}

	mongoDB, err := mongo.NewMongoClient(ctx, cfg.MongoClientConfig)
	if err != nil {
		slog.Error("Failed to create MongoDB client", "error", err)
		return
	}

	repo, err := repository.NewMongoSummaryRepository(mongoDB)
	if err != nil {
		slog.Error("Failed to create summary repository", "error", err)
		return
	}

	// Create posters for different social media
	poster, err := initializePoster(
		ctx,
		cfg.ThreadsClientConfig,
		cfg.TwitterClientConfig,
		cfg.KafkaReaderConfig,
		httpClient,
		repo,
	)
	if err != nil {
		slog.Error("Failed to initialize poster service", "error", err)
		return
	}

	// Run the poster in background
	if err := startBackgroundJobs(ctx, poster); err != nil {
		slog.Error("Failed to run background jobs concurrently", "error", err)
		return
	}

	// Perform a safe shutdown
	if err := safeShutDown(ctx, poster, mongoDB); err != nil {
		slog.Error("Failed to perform graceful shutdown", "error", err)
		return
	}

	slog.Info("Flight Poster service has fully stopped")
}

func initializePoster(
	ctx context.Context,
	threadsCfg config.ThreadsAPIConfig,
	twittercfg config.TwitterAPIConfig,
	kafkaReaderCfg kafka.ReaderConfig,
	httpClient *http.Client,
	repo repository.SummaryRepository,
) (*service.Poster, error) {
	// Create posters for different social media
	threads, err := client.NewThreadsAPI(ctx, threadsCfg, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create Threads client: %w", err)
	}

	twitter, err := client.NewTwitterAPI(twittercfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Twitter client: %w", err)
	}

	kafkaReader, err := kafka.NewKafkaReader(kafkaReaderCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka reader: %w", err)
	}

	clients := []client.Socials{threads, twitter}
	poster, err := service.NewPoster(clients, kafkaReader, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create poster service: %w", err)
	}

	return poster, nil
}

// startBackgroundJobs starts the processor service in background.
func startBackgroundJobs(ctx context.Context, poster *service.Poster) error {
	// Use errgroup to manage concurrent tasks
	g, gCtx := errgroup.WithContext(ctx)

	// Start the processor service
	g.Go(func() error {
		return poster.Post(gCtx)
	})

	// Wait for the context to be done (e.g., due to an interrupt signal)
	<-gCtx.Done()

	// Wait for all goroutines to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to run poster: %w", err)
	}

	return nil
}

// safeShutDown shut down MongoDB client and kafka reader gracefully.
func safeShutDown(ctx context.Context, poster *service.Poster, mongodb *mongo.Client) error {
	if err := mongodb.Client.Disconnect(ctx); err != nil {
		slog.Error("Failed to shutdown MongoDB client", "error", err)
		return fmt.Errorf("failed to shutdown mongodb client: %w", err)
	}

	poster.Close()

	return nil
}

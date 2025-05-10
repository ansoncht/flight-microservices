package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansoncht/flight-microservices/internal/processor/config"
	"github.com/ansoncht/flight-microservices/internal/processor/repository"
	"github.com/ansoncht/flight-microservices/internal/processor/service"

	"github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/ansoncht/flight-microservices/pkg/logger"
	"github.com/ansoncht/flight-microservices/pkg/mongo"
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

	// Create processor service to gather statistic
	processor, err := initializeProcessorService(
		cfg.KafkaWriterConfig,
		cfg.KafkaReaderConfig,
		cfg.SummarizerConfig,
		repo,
	)
	if err != nil {
		slog.Error("Failed to initialize processor service", "error", err)
		return
	}

	// Run the processor in background
	if err := startBackgroundJobs(ctx, processor); err != nil {
		slog.Error("Failed to run background jobs concurrently", "error", err)
		return
	}

	// Perform a safe shutdown
	if err := safeShutDown(ctx, processor, mongoDB); err != nil {
		slog.Error("Failed to perform graceful shutdown", "error", err)
		return
	}

	slog.Info("Flight Processor service has fully stopped")
}

// initializeReaderService initializes the processor service.
func initializeProcessorService(
	kafkaWriterCfg kafka.WriterConfig,
	kafkaReaderCfg kafka.ReaderConfig,
	summarizerCfg config.SummarizerConfig,
	repo repository.SummaryRepository,
) (*service.Processor, error) {
	kafkaWriter, err := kafka.NewKafkaWriter(kafkaWriterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka writer: %w", err)
	}

	kafkaReader, err := kafka.NewKafkaReader(kafkaReaderCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka reader: %w", err)
	}

	summarizer, err := service.NewSummarizer(summarizerCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create summarizer: %w", err)
	}

	processor, err := service.NewProcessor(kafkaWriter, kafkaReader, summarizer, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create processor service: %w", err)
	}

	return processor, nil
}

// startBackgroundJobs starts the processor service in background.
func startBackgroundJobs(ctx context.Context, processor *service.Processor) error {
	// Use errgroup to manage concurrent tasks
	g, gCtx := errgroup.WithContext(ctx)

	// Start the processor service
	g.Go(func() error {
		return processor.Process(gCtx)
	})

	// Wait for the context to be done (e.g., due to an interrupt signal)
	<-gCtx.Done()

	// Wait for all goroutines to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to run processor: %w", err)
	}

	return nil
}

// safeShutDown shut down MongoDB client and kafka reader gracefully.
func safeShutDown(ctx context.Context, processor *service.Processor, mongodb *mongo.Client) error {
	slog.Info("Shutting down components")

	if err := mongodb.Client.Disconnect(ctx); err != nil {
		slog.Error("Failed to shutdown MongoDB client", "error", err)
		return fmt.Errorf("failed to shutdown mongodb client: %w", err)
	}

	processor.MessageReader.Close()

	return nil
}

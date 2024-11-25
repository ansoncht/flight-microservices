package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const format = "2006-01-02"

type MongoDB struct {
	opts   *options.ClientOptions
	client *mongo.Client
}

func NewMongoDB() (*MongoDB, error) {
	slog.Info("Creating MongoDB connection for the service")

	cfg, err := config.MakeMongoDBConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get mongo db config: %w", err)
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(cfg.URI).SetServerAPIOptions(serverAPI)

	return &MongoDB{
		opts:   opts,
		client: nil,
	}, nil
}

func (m *MongoDB) Connect(ctx context.Context) error {
	slog.Info("Starting MongoDB client connection")

	var err error

	m.client, err = mongo.Connect(ctx, m.opts)
	if err != nil {
		return fmt.Errorf("failed to connect to mongo db: %w", err)
	}

	var result bson.M

	if err := m.client.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		return fmt.Errorf("failed to ping mongo db: %w", err)
	}

	return nil
}

func (m *MongoDB) Disconnect(ctx context.Context) error {
	if m.client == nil {
		slog.Info("No active Mongo DB connection")

		return nil
	}

	if err := m.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from mongo db: %w", err)
	}

	return nil
}

func (m *MongoDB) InsertSummary(ctx context.Context, summary model.FlightSummary, date string) error {
	cfg, err := config.MakeMongoDBConfig()
	if err != nil {
		return fmt.Errorf("failed to get mongo db config: %w", err)
	}

	dt, _ := time.Parse(format, date)
	summary.Date = primitive.NewDateTimeFromTime(dt)

	if _, err := m.client.Database(cfg.DB).Collection(cfg.COLLECTION).InsertOne(ctx, summary); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			slog.Warn("Skipping insertion with the same date")

			return nil
		}

		return fmt.Errorf("failed to store daily summary in mongo db: %w", err)
	}

	slog.Info("Stored flight summary", "date", date, "summary", summary.Summary)

	return nil
}

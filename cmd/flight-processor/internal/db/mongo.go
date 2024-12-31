package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const format = "2006-01-02"

// HTTP struct represents the Mongo client and its dependencies.
type Mongo struct {
	opts   *options.ClientOptions // options for Mongo connetion
	client *mongo.Client          // client communicating with the Mongo DB
}

// NewMongo creates a new Mongo client instance.
func NewMongo() (*Mongo, error) {
	slog.Info("Creating Mongo client for the service")

	cfg, err := config.LoadMongoClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load mongo client config: %w", err)
	}

	// Apply the URI from the configuration to the client options
	opts := options.Client().ApplyURI(cfg.URI)

	return &Mongo{
		opts:   opts,
		client: nil,
	}, nil
}

// Connect establishes a connection to the MongoDB server.
func (m *Mongo) Connect(ctx context.Context) error {
	slog.Info("Starting Mongo client connection")

	client, err := mongo.Connect(ctx, m.opts)
	if err != nil {
		return fmt.Errorf("failed to connect to mongo: %w", err)
	}

	m.client = client

	// Ping the primary server to verify the connection
	if err := m.client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping mongo: %w", err)
	}

	return nil
}

// Disconnect closes the connection to the MongoDB server.
func (m *Mongo) Disconnect(ctx context.Context) error {
	if err := m.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from mongo: %w", err)
	}

	m.client = nil

	return nil
}

// InsertSummary stores a flight summary for a specific date in the MongoDB collection.
func (m *Mongo) InsertSummary(ctx context.Context, data map[string]int, date string) error {
	cfg, err := config.LoadMongoClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load mongo client config: %w", err)
	}

	// Parse the provided date string into a time.Time object
	dt, err := time.Parse(format, date)
	if err != nil {
		return fmt.Errorf("failed to parse date for transaction: %w", err)
	}

	// Set the object to be inserted
	summary := &model.FlightSummary{
		Date:    primitive.NewDateTimeFromTime(dt),
		Summary: data,
	}

	collection := m.client.Database(cfg.DB).Collection("dailySummaries")

	// Attempt to insert the summary into the collection
	if _, err := collection.InsertOne(ctx, summary); err != nil {
		// Check for duplicate key error and log a warning
		if mongo.IsDuplicateKeyError(err) {
			slog.Warn("Skipping insertion with the same date")
			return nil
		}

		return fmt.Errorf("failed to store daily summary in mongo: %w", err)
	}

	slog.Info("Stored flight summary", "date", date, "summary", summary.Summary)

	return nil
}

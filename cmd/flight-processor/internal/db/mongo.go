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

type Database interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	InsertSummary(ctx context.Context, data map[string]int, date string) error
}

// Mongo holds the Mongo client and its configuration settings.
type Mongo struct {
	// opts specifies the client options for connecting to MongoDB.
	opts *options.ClientOptions
	// client is the MongoDB client used to communicate with the database.
	client *mongo.Client
	// db specifies the name of the MongoDB database to use.
	db string
}

// NewMongo creates a new MongoDB client based on the provided configuration.
func NewMongo(cfg config.MongoClientConfig) (*Mongo, error) {
	slog.Info("Creating Mongo client for the service")

	if cfg.URI == "" {
		return nil, fmt.Errorf("MongoDB connection URI is empty")
	}
	if cfg.DB == "" {
		return nil, fmt.Errorf("MongoDB database name is empty")
	}

	// Apply the URI from the configuration to the client options.
	opts := options.Client().ApplyURI(cfg.URI)

	return &Mongo{
		opts:   opts,
		client: nil,
		db:     cfg.DB,
	}, nil
}

// Connect establishes a connection to the MongoDB server and verifies it.
func (m *Mongo) Connect(ctx context.Context) error {
	slog.Info("Starting MongoDB client connection")

	client, err := mongo.Connect(ctx, m.opts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	m.client = client

	// Ping the primary server to verify the connection.
	if err := m.client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	slog.Info("Successfully connected to MongoDB")

	return nil
}

// Disconnect closes the MongoDB client connection.
func (m *Mongo) Disconnect(ctx context.Context) error {
	if err := m.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	m.client = nil

	return nil
}

// InsertSummary adds a flight summary for a specific date to the MongoDB collection.
func (m *Mongo) InsertSummary(ctx context.Context, data map[string]int, date string) error {
	dt, err := parseDate(date)
	if err != nil {
		return fmt.Errorf("failed to parse date for transaction: %w", err)
	}

	// Prepare the summary object to be inserted.
	summary := &model.FlightSummary{
		Date:    primitive.NewDateTimeFromTime(dt),
		Summary: data,
	}

	collection := m.client.Database(m.db).Collection("dailySummaries")

	// Attempt to insert the summary into the collection.
	if _, err := collection.InsertOne(ctx, summary); err != nil {
		// Log a warning if a duplicate key error occurs.
		if mongo.IsDuplicateKeyError(err) {
			slog.Warn("Skipping insertion with the same date")
			return nil
		}

		return fmt.Errorf("failed to store daily summary in MongoDB: %w", err)
	}

	slog.Info("Stored flight summary", "date", date, "summary", summary.Summary)

	return nil
}

// parseDate parses the date string into a time.Time object.
func parseDate(date string) (time.Time, error) {
	dt, err := time.Parse(format, date)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date: %w", err)
	}

	return dt, nil
}

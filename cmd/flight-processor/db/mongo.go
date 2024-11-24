package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func (c *MongoDB) Connect(ctx context.Context) error {
	slog.Info("Starting MongoDB client connection")

	var err error

	c.client, err = mongo.Connect(ctx, c.opts)
	if err != nil {
		return fmt.Errorf("failed to connect to mongo db: %w", err)
	}

	var result bson.M

	if err := c.client.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Decode(&result); err != nil {
		return fmt.Errorf("failed to ping mongo db: %w", err)
	}

	return nil
}

func (c *MongoDB) Disconnect(ctx context.Context) error {
	if c.client == nil {
		slog.Info("No active Mongo DB connection")

		return nil
	}

	if err := c.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from mongo db: %w", err)
	}

	return nil
}

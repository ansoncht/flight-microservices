package mongo

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ClientConfig holds configuration settings for the MongoDB client.
type ClientConfig struct {
	// URI specifies the connection URI for the MongoDB server.
	URI string `mapstructure:"uri"`
	// DB specifies the name of the MongoDB database to use.
	DB string `mapstructure:"db"`
	// PoolSize specifies the maximum number of connections in the connection pool.
	PoolSize uint64 `mapstructure:"pool_size"`
	// ConnectionTimeout specifies the maximum time to wait for a connection.
	ConnectionTimeout int `mapstructure:"connection_timeout"`
	// SocketTimeout specifies the maximum time to wait for a socket operation.
	SocketTimeout int `mapstructure:"socket_timeout"`
}

// Client holds the MongoDB client instance and the database to use.
type Client struct {
	// Client specifies the MongoDB client used to communicate with the database.
	Client *mongo.Client
	// Database specifies the MongoDB database to use.
	Database *mongo.Database
}

// NewMongoClient creates a new MongoClient instance based on the provided configuration.
func NewMongoClient(ctx context.Context, cfg ClientConfig) (*Client, error) {
	slog.Info("Initializing MongoDB client for the service", "uri", cfg.URI)

	if cfg.URI == "" {
		return nil, fmt.Errorf("MongoDB connection URI is empty")
	}

	if cfg.DB == "" {
		return nil, fmt.Errorf("MongoDB database name is empty")
	}

	if cfg.PoolSize <= 0 {
		return nil, fmt.Errorf("MongoDB pool size is invalid: %d", cfg.PoolSize)
	}

	if cfg.ConnectionTimeout <= 0 {
		return nil, fmt.Errorf("MongoDB connection timeout is invalid: %d", cfg.ConnectionTimeout)
	}

	if cfg.SocketTimeout <= 0 {
		return nil, fmt.Errorf("MongoDB socket timeout is invalid: %d", cfg.SocketTimeout)
	}

	clientOpts := options.Client().ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.PoolSize).
		SetConnectTimeout(time.Duration(cfg.ConnectionTimeout) * time.Second).
		SetSocketTimeout(time.Duration(cfg.SocketTimeout) * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the primary server to verify the connection.
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(cfg.ConnectionTimeout)*time.Second)
	defer cancel()
	if err := client.Ping(ctxTimeout, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(cfg.DB)
	return &Client{
		Client:   client,
		Database: db,
	}, nil
}

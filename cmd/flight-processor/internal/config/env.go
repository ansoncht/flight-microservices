package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

// GRPCServerConfig represents gRPC server configurations.
type GRPCServerConfig struct {
	PORT int `env:"SERVER_GRPC_PORT"`
}

// MongoDBConfig represents Mongo DB configurations.
type MongoDBConfig struct {
	URI        string `env:"MONGO_DB_URI"`
	DB         string `env:"MONGO_DB_NAME"`
	COLLECTION string `env:"MONGO_DB_COLLECTION"`
}

// MakeGRPCServerConfig parses environment variables into a GRPCServerConfig struct.
func MakeGRPCServerConfig() (*GRPCServerConfig, error) {
	var cfg GRPCServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for gRPC server config: %w", err)
	}

	return &cfg, nil
}

// MakeMongoDBConfig parses environment variables into a MongoDBConfig struct.
func MakeMongoDBConfig() (*MongoDBConfig, error) {
	var cfg MongoDBConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for mongo db config: %w", err)
	}

	return &cfg, nil
}

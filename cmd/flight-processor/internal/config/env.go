package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

// GrpcServerConfig represents the configuration for the gRPC server.
type GrpcServerConfig struct {
	Port string `env:"FLIGHT_PROCESSOR_GRPC_PORT"`
}

// GrpcClientConfig represents the configuration for the gRPC client.
type GrpcClientConfig struct {
	Address string `env:"FLIGHT_PROCESSOR_GRPC_ADDRESS"` // Address of the gRPC server which gRPC client connects
}

// MongoClientConfig represents the configuration for the Mongo client.
type MongoClientConfig struct {
	URI string `env:"FLIGHT_PROCESSOR_MONGO_URI"`
	DB  string `env:"FLIGHT_PROCESSOR_MONGO_DB"`
}

// LoadGrpcServerConfig parses environment variables into a GrpcServerConfig struct.
func LoadGrpcServerConfig() (*GrpcServerConfig, error) {
	var cfg GrpcServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for gRPC server config: %w", err)
	}

	return &cfg, nil
}

// LoadGrpcClientConfig parses environment variables into a GRPCClientConfig struct.
func LoadGrpcClientConfig() (*GrpcClientConfig, error) {
	var cfg GrpcClientConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for gRPC client config: %w", err)
	}

	return &cfg, nil
}

// LoadMongoClientConfig parses environment variables into a MongoClientConfig struct.
func LoadMongoClientConfig() (*MongoClientConfig, error) {
	var cfg MongoClientConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for mongo db config: %w", err)
	}

	return &cfg, nil
}

package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

// GRPCServerConfig represents gRPC server configurations.
type GRPCServerConfig struct {
	PORT int `env:"SERVER_GRPC_PORT"`
}

// MakeGRPCServerConfig parses environment variables into a GRPCServerConfig struct.
func MakeGRPCServerConfig() (*GRPCServerConfig, error) {
	var cfg GRPCServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for gRPC server config: %w", err)
	}

	if cfg.PORT <= 0 {
		return nil, fmt.Errorf("invalid port number: %d, must be positive number", cfg.PORT)
	}

	return &cfg, nil
}

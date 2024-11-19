package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

// GRPCServerConfig represents gRPC server configurations.
type GRPCServerConfig struct {
	PORT int `env:"SERVER_GRPC_PORT"`
}

func MakeGRPCServerConfig() (*GRPCServerConfig, error) {
	var cfg GRPCServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for gRPC server config: %w", err)
	}

	return &cfg, nil
}

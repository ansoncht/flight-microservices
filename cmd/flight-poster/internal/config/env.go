package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

// GrpcServerConfig represents gRPC server configurations.
type GrpcServerConfig struct {
	Port string `env:"FLIGHT_POSTER_GRPC_PORT"`
}

// ThreadsClientConfig represents Threads client configurations.
type ThreadsClientConfig struct {
	URL   string `env:"FLIGHT_POSTER_THREADS_URL"`
	Token string `env:"FLIGHT_POSTER_THREADS_TOKEN"`
}

// HTTPClientConfig represents the configuration for the HTTP client.
type HTTPClientConfig struct {
	Timeout int64 `env:"FLIGHT_POSTER_HTTP_TIMEOUT"` // Timeout for reading HTTP headers in seconds
}

// LoadGrpcServerConfig parses environment variables into a GrpcServerConfig struct.
func LoadGrpcServerConfig() (*GrpcServerConfig, error) {
	var cfg GrpcServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for gRPC server config: %w", err)
	}

	return &cfg, nil
}

// LoadThreadsClientConfig parses environment variables into a ThreadsClientConfig struct.
func LoadThreadsClientConfig() (*ThreadsClientConfig, error) {
	var cfg ThreadsClientConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for Threads client config: %w", err)
	}

	return &cfg, nil
}

// LoadHTTPClientConfig parses environment variables into a HTTPClientConfig struct.
func LoadHTTPClientConfig() (*HTTPClientConfig, error) {
	var cfg HTTPClientConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for HTTP client config: %w", err)
	}

	return &cfg, nil
}

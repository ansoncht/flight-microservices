package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

// LoggerConfig represents logger configurations.
type LoggerConfig struct {
	JSON  bool   `env:"LOGGER_JSON"`
	Level string `env:"LOGGER_LEVEL"`
}

// HTTPConfig represents http server configurations.
type HTTPServerConfig struct {
	Port              string `env:"SERVER_HTTP_PORT"`
	ReadHeaderTimeout int64  `env:"SERVER_HTTP_READ_TIMEOUT"`
}

// HTTPClientConfig represents http client configurations.
type HTTPClientConfig struct {
	URL     string `env:"CLIENT_HTTP_FETCH_URL"`
	Timeout int64  `env:"CLIENT_HTTP_TIMEOUT"`
}

// GRPCClientConfig represents gRPC client configurations.
type GRPCClientConfig struct {
	URL string `env:"CLIENT_GRPC_URL"`
}

// MakeLoggerConfig parses environment variables into a LoggerConfig struct.
func MakeLoggerConfig() (*LoggerConfig, error) {
	var cfg LoggerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for logger config: %w", err)
	}

	return &cfg, nil
}

// MakeHTTPServerConfig parses environment variables into a HTTPServerConfig struct.
func MakeHTTPServerConfig() (*HTTPServerConfig, error) {
	var cfg HTTPServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for http server config: %w", err)
	}

	return &cfg, nil
}

// MakeHTTPClientConfig parses environment variables into a HTTPClientConfig struct.
func MakeHTTPClientConfig() (*HTTPClientConfig, error) {
	var cfg HTTPClientConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for http client config: %w", err)
	}

	return &cfg, nil
}

// MakeGRPCClientConfig parses environment variables into a GRPCClientConfig struct.
func MakeGRPCClientConfig() (*GRPCClientConfig, error) {
	var cfg GRPCClientConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for gRPC client config: %w", err)
	}

	return &cfg, nil
}

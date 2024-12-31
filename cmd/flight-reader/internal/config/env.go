package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

// HTTPServerConfig represents the configuration for the HTTP server.
type HTTPServerConfig struct {
	Port    string `env:"FLIGHT_READER_HTTP_PORT"`    // Port on which the HTTP server listens
	Timeout int64  `env:"FLIGHT_READER_HTTP_TIMEOUT"` // Timeout for reading HTTP headers in seconds
}

// HTTPClientConfig represents the configuration for the HTTP client.
type HTTPClientConfig struct {
	Timeout int64 `env:"FLIGHT_READER_HTTP_TIMEOUT"` // Timeout for reading HTTP headers in seconds
}

// FlightFetcherConfig represents the configuration for the flight fetcher.
type FlightFetcherConfig struct {
	URL  string `env:"FLIGHT_READER_FLIGHT_URL"`  // Base URL for the flight fetcher
	User string `env:"FLIGHT_READER_FLIGHT_USER"` // Username for accessing the API
	Pass string `env:"FLIGHT_READER_FLIGHT_PASS"` // Password for accessing the API
}

// RouteFetcherConfig represents the configuration for the route fetcher.
type RouteFetcherConfig struct {
	URL string `env:"FLIGHT_READER_ROUTE_URL"` // Base URL for the route fetcher
}

// GrpcClientConfig represents the configuration for the gRPC client.
type GrpcClientConfig struct {
	Address string `env:"FLIGHT_READER_GRPC_ADDRESS"` // Address of the gRPC server which gRPC client connects
}

// SchedulerConfig represents the configuration for the scheduler.
type SchedulerConfig struct {
	Airports string `env:"FLIGHT_READER_SCHEDULER_AIRPORTS"` // Airports which the scheduler fetches data for
}

// LoadHTTPServerConfig parses environment variables into a HTTPServerConfig struct.
func LoadHTTPServerConfig() (*HTTPServerConfig, error) {
	var cfg HTTPServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for http server config: %w", err)
	}

	return &cfg, nil
}

// LoadHTTPClientConfig parses environment variables into a HTTPServerConfig struct.
func LoadHTTPClientConfig() (*HTTPClientConfig, error) {
	var cfg HTTPClientConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for http server config: %w", err)
	}

	return &cfg, nil
}

// LoadFlightFetcherConfig parses environment variables into a FlightFetcherConfig struct.
func LoadFlightFetcherConfig() (*FlightFetcherConfig, error) {
	var cfg FlightFetcherConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for flight fetcher config: %w", err)
	}

	return &cfg, nil
}

// LoadRouteFetcherConfig parses environment variables into a RouteFetcherConfig struct.
func LoadRouteFetcherConfig() (*RouteFetcherConfig, error) {
	var cfg RouteFetcherConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for route fetcher config: %w", err)
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

// LoadSchedulerConfig parses environment variables into a SchedulerConfig struct.
func LoadSchedulerConfig() (*SchedulerConfig, error) {
	var cfg SchedulerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for scheduler config: %w", err)
	}

	return &cfg, nil
}

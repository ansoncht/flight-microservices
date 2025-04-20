package config

import (
	"fmt"
	"strings"

	"github.com/ansoncht/flight-microservices/pkg/logger"
	"github.com/spf13/viper"
)

// FlightReaderConfig holds all configurations related to flight reader.
type FlightReaderConfig struct {
	HTTPServerConfig    HTTPServerConfig    `mapstructure:"http-server"`
	HTTPClientConfig    HTTPClientConfig    `mapstructure:"http-client"`
	GrpcClientConfig    GrpcClientConfig    `mapstructure:"grpc-client"`
	FlightFetcherConfig FlightFetcherConfig `mapstructure:"flight-fetcher"`
	RouteFetcherConfig  RouteFetcherConfig  `mapstructure:"route-fetcher"`
	SchedulerConfig     SchedulerConfig     `mapstructure:"scheduler"`
	LoggerConfig        logger.Config       `mapstructure:"logger"`
}

// HTTPServerConfig holds configuration settings for the HTTP server.
type HTTPServerConfig struct {
	// Port specifies the port where the HTTP server listens for requests.
	Port string `mapstructure:"port"`
	// Timeout specifies the timeout for reading HTTP headers in seconds.
	Timeout int `mapstructure:"timeout"`
}

// HTTPClientConfig holds configuration settings for the HTTP client.
type HTTPClientConfig struct {
	// Timeout specifies the timeout for reading HTTP headers in seconds.
	Timeout int `mapstructure:"timeout"`
}

// GrpcClientConfig holds configuration settings for the gRPC client.
type GrpcClientConfig struct {
	// Address specifies the address of the Processor gRPC server.
	Address string `mapstructure:"address"`
}

// FlightFetcherConfig holds configuration settings for the flight fetcher.
type FlightFetcherConfig struct {
	// URL specifies the base URL for the flight fetcher.
	URL string `mapstructure:"url"`
	// User specifies the username for accessing the API.
	User string `mapstructure:"user"`
	// Pass specifies the password for accessing the API.
	Pass string `mapstructure:"pass"`
}

// RouteFetcherConfig holds configuration settings for the route fetcher.
type RouteFetcherConfig struct {
	// URL specifies the base URL for the flight fetcher.
	URL string `mapstructure:"url"`
}

// SchedulerConfig holds configuration settings for the scheduler.
type SchedulerConfig struct {
	// Airports specifies the airports for which the scheduler fetches data.
	Airports string `mapstructure:"airports"`
}

// LoadConfig loads configuration from environment variables and a YAML file.
func LoadConfig() (*FlightReaderConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../../")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("FLIGHT_READER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg FlightReaderConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

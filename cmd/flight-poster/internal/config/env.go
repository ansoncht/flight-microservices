package config

import (
	"fmt"
	"strings"

	"github.com/ansoncht/flight-microservices/pkg/logger"
	"github.com/spf13/viper"
)

// FlightPosterConfig holds all configurations related to flight poster.
type FlightPosterConfig struct {
	GrpcServerConfig    GrpcServerConfig    `mapstructure:"grpc-server"`
	HTTPClientConfig    HTTPClientConfig    `mapstructure:"http-client"`
	ThreadsClientConfig ThreadsClientConfig `mapstructure:"threads"`
	TwitterClientConfig TwitterClientConfig `mapstructure:"twitter"`
	LoggerConfig        logger.Config       `mapstructure:"logger"`
}

// GrpcServerConfig holds configuration settings for the gRPC server.
type GrpcServerConfig struct {
	// Port specifies the port where the gRPC server listens for connections.
	Port string `mapstructure:"port"`
}

// HTTPClientConfig holds configuration settings for the HTTP client.
type HTTPClientConfig struct {
	// Timeout specifies the timeout for reading HTTP headers in seconds.
	Timeout int `mapstructure:"timeout"`
}

// ThreadsClientConfig holds configuration settings for the Threads client.
type ThreadsClientConfig struct {
	// URL specifies the base URL for the Threads API.
	URL string `mapstructure:"url"`
	// Token specifies the access token for authentication with the Threads API.
	Token string `mapstructure:"access-token"`
}

// TwitterClientConfig holds configuration settings for the Twitter client.
type TwitterClientConfig struct {
	// Key specifies the API key for Twitter authentication.
	Key string `mapstructure:"access-token-key"`
	// Secret specifies the API secret key for Twitter authentication.
	Secret string `mapstructure:"access-token-secret"`
}

// LoadConfig loads configuration from environment variables and a YAML file.
func LoadConfig() (*FlightPosterConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("FLIGHT_POSTER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg FlightPosterConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

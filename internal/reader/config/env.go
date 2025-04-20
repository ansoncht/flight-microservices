package config

import (
	"fmt"
	"strings"

	"github.com/ansoncht/flight-microservices/pkg/client"
	"github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/ansoncht/flight-microservices/pkg/logger"
	"github.com/ansoncht/flight-microservices/pkg/server"
	"github.com/spf13/viper"
)

// FlightReaderConfig holds all configurations related to flight reader.
type FlightReaderConfig struct {
	HTTPServerConfig      server.HTTPConfig     `mapstructure:"http_server"`
	HTTPClientConfig      client.HTTPConfig     `mapstructure:"http_client"`
	FlightAPIClientConfig FlightAPIClientConfig `mapstructure:"flight_api_client"`
	RouteAPIClientConfig  RouteAPIClientConfig  `mapstructure:"route_api_client"`
	KafkaWriterConfig     kafka.WriterConfig    `mapstructure:"kafka_writer"`
	LoggerConfig          logger.Config         `mapstructure:"logger"`
}

// FlightAPIClientConfig holds configuration settings for the flight api client.
type FlightAPIClientConfig struct {
	// URL specifies the base URL for the flight api.
	URL string `mapstructure:"url"`
	// User specifies the username for accessing the API.
	User string `mapstructure:"user"`
	// Pass specifies the password for accessing the API.
	Pass string `mapstructure:"pass"`
}

// RouteAPIClientConfig holds configuration settings for the route api client.
type RouteAPIClientConfig struct {
	// URL specifies the base URL for the route api.
	URL string `mapstructure:"url"`
}

// LoadConfig loads configuration from environment variables and a YAML file.
func LoadConfig() (*FlightReaderConfig, error) {
	viper.SetConfigName("reader-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../configs")
	viper.AddConfigPath("../../../configs")
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

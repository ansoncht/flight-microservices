package config

import (
	"fmt"
	"strings"

	"github.com/ansoncht/flight-microservices/pkg/http"
	"github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/ansoncht/flight-microservices/pkg/logger"
	"github.com/ansoncht/flight-microservices/pkg/mongo"
	"github.com/spf13/viper"
)

// FlightPosterConfig holds all configurations related to flight poster.
type FlightPosterConfig struct {
	ThreadsClientConfig ThreadsAPIConfig   `mapstructure:"threads_api"`
	TwitterClientConfig TwitterAPIConfig   `mapstructure:"twitter_api"`
	KafkaReaderConfig   kafka.ReaderConfig `mapstructure:"kafka_reader"`
	MongoClientConfig   mongo.ClientConfig `mapstructure:"mongo"`
	HTTPClientConfig    http.ClientConfig  `mapstructure:"http_client"`
	LoggerConfig        logger.Config      `mapstructure:"logger"`
}

// ThreadsAPIConfig holds configuration settings for the Threads api client.
type ThreadsAPIConfig struct {
	// URL specifies the base URL for the Threads API.
	URL string `mapstructure:"url"`
	// Token specifies the access token for authentication with the Threads API.
	Token string `mapstructure:"access_token"`
}

// TwitterAPIConfig holds configuration settings for the Twitter api client.
type TwitterAPIConfig struct {
	// Key specifies the API key for Twitter authentication.
	Key string `mapstructure:"access_token_key"`
	// Secret specifies the API secret key for Twitter authentication.
	Secret string `mapstructure:"access_token_secret"`
}

// LoadConfig loads configuration from environment variables and a YAML file.
func LoadConfig() (*FlightPosterConfig, error) {
	viper.SetConfigName("poster-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../../configs")
	viper.AddConfigPath("../../configs")
	viper.AddConfigPath("./configs")
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

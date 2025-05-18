package config

import (
	"fmt"
	"strings"

	"github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/ansoncht/flight-microservices/pkg/logger"
	"github.com/ansoncht/flight-microservices/pkg/mongo"
	"github.com/spf13/viper"
)

// FlightProcessorConfig holds all configurations related to flight processor.
type FlightProcessorConfig struct {
	SummarizerConfig  SummarizerConfig   `mapstructure:"summarizer"`
	MongoClientConfig mongo.ClientConfig `mapstructure:"mongo"`
	KafkaWriterConfig kafka.WriterConfig `mapstructure:"kafka_writer"`
	KafkaReaderConfig kafka.ReaderConfig `mapstructure:"kafka_reader"`
	LoggerConfig      logger.Config      `mapstructure:"logger"`
}

// SummarizerConfig holds configuration settings for the summarizer.
type SummarizerConfig struct {
	// TopN specifies the number of top airlines and destinations to summarize.
	TopN int `mapstructure:"top_n"`
}

// LoadConfig loads configuration from environment variables and a YAML file.
func LoadConfig() (*FlightProcessorConfig, error) {
	viper.SetConfigName("processor-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../../configs")
	viper.AddConfigPath("../../configs")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("FLIGHT_PROCESSOR")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg FlightProcessorConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

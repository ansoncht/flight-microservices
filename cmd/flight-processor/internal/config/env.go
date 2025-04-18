package config

import (
	"fmt"
	"strings"

	"github.com/ansoncht/flight-microservices/pkg/logger"
	"github.com/spf13/viper"
)

// FlightProcessorConfig holds all configurations related to flight processor.
type FlightProcessorConfig struct {
	GrpcServerConfig  GrpcServerConfig  `mapstructure:"grpc-server"`
	GrpcClientConfig  GrpcClientConfig  `mapstructure:"grpc-client"`
	MongoClientConfig MongoClientConfig `mapstructure:"mongodb"`
	LoggerConfig      logger.Config     `mapstructure:"logger"`
}

// GrpcServerConfig holds configuration settings for the gRPC server.
type GrpcServerConfig struct {
	// Port specifies the port where the gRPC server listens for connections.
	Port string `mapstructure:"port"`
}

// GrpcClientConfig holds configuration settings for the gRPC client.
type GrpcClientConfig struct {
	// Address specifies the address of the Poster gRPC server.
	Address string `mapstructure:"address"`
}

// MongoClientConfig holds configuration settings for the MongoDB client.
type MongoClientConfig struct {
	// URI specifies the connection URI for the MongoDB server.
	URI string `mapstructure:"uri"`
	// DB specifies the name of the MongoDB database to use.
	DB string `mapstructure:"db"`
}

// LoadConfig loads configuration from environment variables and a YAML file.
func LoadConfig() (*FlightProcessorConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
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

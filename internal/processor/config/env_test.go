package config_test

import (
	"os"
	"testing"

	"github.com/ansoncht/flight-microservices/internal/processor/config"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_ValidConfigFile_ShouldSucceed(t *testing.T) {
	os.Setenv("FLIGHT_PROCESSOR_MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("FLIGHT_PROCESSOR_KAFKA_READER_ADDRESS", "test")
	os.Setenv("FLIGHT_PROCESSOR_KAFKA_READER_TOPIC", "test")
	os.Setenv("FLIGHT_PROCESSOR_KAFKA_READER_GROUP_ID", "test")
	os.Setenv("FLIGHT_PROCESSOR_KAFKA_WRITER_ADDRESS", "test")
	os.Setenv("FLIGHT_PROCESSOR_KAFKA_WRITER_TOPIC", "test")

	cfg, err := config.LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "mongodb://localhost:27017", cfg.MongoClientConfig.URI)
	require.Equal(t, "flights", cfg.MongoClientConfig.DB)
	require.Equal(t, uint64(5), cfg.MongoClientConfig.PoolSize)
	require.Equal(t, 5, cfg.MongoClientConfig.ConnectionTimeout)
	require.Equal(t, 5, cfg.MongoClientConfig.SocketTimeout)
	require.Equal(t, "test", cfg.KafkaWriterConfig.Address)
	require.Equal(t, "test", cfg.KafkaWriterConfig.Topic)
	require.Equal(t, "test", cfg.KafkaReaderConfig.Address)
	require.Equal(t, "test", cfg.KafkaReaderConfig.Topic)
	require.Equal(t, "test", cfg.KafkaReaderConfig.GroupID)
	require.True(t, cfg.LoggerConfig.JSON)
	require.Equal(t, "info", cfg.LoggerConfig.Level)
}

func TestLoadConfig_MissingFile_ShouldError(t *testing.T) {
	// Temporarily rename the config file if it exists
	originalPath := "../../../configs/processor-config.yaml"
	tempPath := "../../../configs/processor-config.yaml.bak"

	// Restore the file after the test
	if _, err := os.Stat(originalPath); err == nil {
		err := os.Rename(originalPath, tempPath)
		require.NoError(t, err)
		defer func() {
			err := os.Rename(tempPath, originalPath)
			require.NoError(t, err, "failed to restore config file")
		}()
	}

	cfg, err := config.LoadConfig()
	require.Error(t, err)
	require.Nil(t, cfg)
}

func TestLoadConfig_EnvOverride_ShouldSucceed(t *testing.T) {
	t.Run("Override Config File", func(t *testing.T) {
		os.Setenv("FLIGHT_PROCESSOR_MONGO_URI", "mongodb://localhost:37017")
		os.Setenv("FLIGHT_PROCESSOR_MONGO_DB", "test_db")
		os.Setenv("FLIGHT_PROCESSOR_LOGGER_LEVEL", "debug")
		os.Setenv("FLIGHT_PROCESSOR_LOGGER_JSON", "false")

		cfg, err := config.LoadConfig()

		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, "mongodb://localhost:37017", cfg.MongoClientConfig.URI)
		require.Equal(t, "test_db", cfg.MongoClientConfig.DB)
		require.False(t, cfg.LoggerConfig.JSON)
		require.Equal(t, "debug", cfg.LoggerConfig.Level)
	})
}

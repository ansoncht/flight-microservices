package config_test

import (
	"os"
	"testing"

	"github.com/ansoncht/flight-microservices/internal/poster/config"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_ValidConfigFile_ShouldSucceed(t *testing.T) {
	os.Setenv("GOTWI_API_KEY", "test")
	os.Setenv("GOTWI_API_KEY_SECRET", "test")
	os.Setenv("FLIGHT_POSTER_THREADS_API_ACCESS_TOKEN", "test")
	os.Setenv("FLIGHT_POSTER_TWITTER_API_ACCESS_TOKEN_KEY", "test")
	os.Setenv("FLIGHT_POSTER_TWITTER_API_ACCESS_TOKEN_SECRET", "test")
	os.Setenv("FLIGHT_POSTER_KAFKA_READER_ADDRESS", "test")
	os.Setenv("FLIGHT_POSTER_KAFKA_READER_TOPIC", "test")
	os.Setenv("FLIGHT_POSTER_KAFKA_READER_GROUP_ID", "test")
	os.Setenv("FLIGHT_POSTER_MONGO_URI", "mongodb://localhost:27017")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, 10, cfg.HTTPClientConfig.Timeout)
	require.Equal(t, "https://graph.threads.net", cfg.ThreadsClientConfig.URL)
	require.Equal(t, "test", cfg.ThreadsClientConfig.Token)
	require.Equal(t, "test", cfg.TwitterClientConfig.Key)
	require.Equal(t, "test", cfg.TwitterClientConfig.Secret)
	require.Equal(t, "mongodb://localhost:27017", cfg.MongoClientConfig.URI)
	require.Equal(t, "flights", cfg.MongoClientConfig.DB)
	require.Equal(t, uint64(5), cfg.MongoClientConfig.PoolSize)
	require.Equal(t, 5, cfg.MongoClientConfig.ConnectionTimeout)
	require.Equal(t, 5, cfg.MongoClientConfig.SocketTimeout)
	require.Equal(t, "test", cfg.KafkaReaderConfig.Address)
	require.Equal(t, "test", cfg.KafkaReaderConfig.Topic)
	require.Equal(t, "test", cfg.KafkaReaderConfig.GroupID)
	require.True(t, cfg.LoggerConfig.JSON)
	require.Equal(t, "info", cfg.LoggerConfig.Level)
}

func TestLoadConfig_MissingFile_ShouldError(t *testing.T) {
	// Temporarily rename the config file if it exists
	originalPath := "../../../configs/poster-config.yaml"
	tempPath := "../../../configs/poster-config.yaml.bak"

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
	os.Setenv("FLIGHT_POSTER_THREADS_API_URL", "test")
	os.Setenv("FLIGHT_POSTER_MONGO_URI", "mongodb://localhost:37017")
	os.Setenv("FLIGHT_POSTER_MONGO_DB", "test_db")
	os.Setenv("FLIGHT_POSTER_LOGGER_LEVEL", "debug")
	os.Setenv("FLIGHT_POSTER_LOGGER_JSON", "false")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "test", cfg.ThreadsClientConfig.URL)
	require.Equal(t, "mongodb://localhost:37017", cfg.MongoClientConfig.URI)
	require.Equal(t, "test_db", cfg.MongoClientConfig.DB)
	require.False(t, cfg.LoggerConfig.JSON)
	require.Equal(t, "debug", cfg.LoggerConfig.Level)
}

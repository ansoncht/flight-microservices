package config_test

import (
	"os"
	"testing"

	"github.com/ansoncht/flight-microservices/internal/reader/config"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_ValidConfigFile_ShouldSucceed(t *testing.T) {
	os.Setenv("FLIGHT_READER_FLIGHT_API_URL", "test")
	os.Setenv("FLIGHT_READER_FLIGHT_API_USER", "test")
	os.Setenv("FLIGHT_READER_FLIGHT_API_PASS", "test")
	os.Setenv("FLIGHT_READER_ROUTE_API_URL", "test")
	os.Setenv("FLIGHT_READER_KAFKA_WRITER_ADDRESS", "test")
	os.Setenv("FLIGHT_READER_KAFKA_WRITER_TOPIC", "test")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "8080", cfg.HTTPServerConfig.Port)
	require.Equal(t, 75, cfg.HTTPServerConfig.Timeout)
	require.Equal(t, 70, cfg.HTTPClientConfig.Timeout)
	require.Equal(t, "test", cfg.FlightAPIClientConfig.URL)
	require.Equal(t, "test", cfg.FlightAPIClientConfig.User)
	require.Equal(t, "test", cfg.FlightAPIClientConfig.Pass)
	require.Equal(t, "test", cfg.KafkaWriterConfig.Address)
	require.Equal(t, "test", cfg.KafkaWriterConfig.Topic)
	require.True(t, cfg.LoggerConfig.JSON)
	require.Equal(t, "info", cfg.LoggerConfig.Level)
}

func TestLoadConfig_MissingFile_ShouldError(t *testing.T) {
	// Temporarily rename the config file if it exists
	originalPath := "../../../configs/reader-config.yaml"
	tempPath := "../../../configs/reader-config.yaml.bak"

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

func TestLoadConfig_UnmarshalError_ShouldError(t *testing.T) {
	originalPath := "../../../configs/reader-config.yaml"
	tempPath := "../../../configs/reader-config.yaml.bak"

	// Backup the original config file if it exists
	if _, err := os.Stat(originalPath); err == nil {
		err := os.Rename(originalPath, tempPath)
		require.NoError(t, err)
		defer func() {
			err := os.Rename(tempPath, originalPath)
			require.NoError(t, err, "failed to restore config file")
		}()
	}

	// Write a config file with a type mismatch (e.g., string instead of int for timeout)
	badConfig := []byte(`
http_client:
  timeout: "not-an-int"
`)
	err := os.WriteFile(originalPath, badConfig, 0600)
	require.NoError(t, err)

	cfg, err := config.LoadConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to unmarshal config")
	require.Nil(t, cfg)
}

func TestLoadConfig_EnvOverride_ShouldSucceed(t *testing.T) {
	os.Setenv("FLIGHT_READER_FLIGHT_API_URL", "test")
	os.Setenv("FLIGHT_READER_FLIGHT_API_USER", "test")
	os.Setenv("FLIGHT_READER_FLIGHT_API_PASS", "test")
	os.Setenv("FLIGHT_READER_ROUTE_API_URL", "test")
	os.Setenv("FLIGHT_READER_KAFKA_WRITER_ADDRESS", "test")
	os.Setenv("FLIGHT_READER_KAFKA_WRITER_TOPIC", "test")
	t.Setenv("FLIGHT_READER_LOGGER_LEVEL", "debug")
	t.Setenv("FLIGHT_READER_LOGGER_JSON", "false")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "8080", cfg.HTTPServerConfig.Port)
	require.Equal(t, 75, cfg.HTTPServerConfig.Timeout)
	require.Equal(t, 70, cfg.HTTPClientConfig.Timeout)
	require.Equal(t, "test", cfg.FlightAPIClientConfig.URL)
	require.Equal(t, "test", cfg.FlightAPIClientConfig.User)
	require.Equal(t, "test", cfg.FlightAPIClientConfig.Pass)
	require.Equal(t, "test", cfg.KafkaWriterConfig.Address)
	require.Equal(t, "test", cfg.KafkaWriterConfig.Topic)
	require.False(t, cfg.LoggerConfig.JSON)
	require.Equal(t, "debug", cfg.LoggerConfig.Level)
}

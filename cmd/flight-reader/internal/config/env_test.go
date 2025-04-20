package config_test

import (
	"os"
	"testing"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_ValidConfigFile_ShouldSucceed(t *testing.T) {
	t.Run("Valid Config File", func(t *testing.T) {
		os.Setenv("FLIGHT_READER_FLIGHT_FETCHER_URL", "test")
		os.Setenv("FLIGHT_READER_FLIGHT_FETCHER_USER", "test")
		os.Setenv("FLIGHT_READER_FLIGHT_FETCHER_PASS", "test")
		os.Setenv("FLIGHT_READER_ROUTE_FETCHER_URL", "test")

		cfg, err := config.LoadConfig()

		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, "9099", cfg.HTTPServerConfig.Port)
		require.Equal(t, 10, cfg.HTTPServerConfig.Timeout)
		require.Equal(t, 10, cfg.HTTPClientConfig.Timeout)
		require.Equal(t, "localhost:9098", cfg.GrpcClientConfig.Address)
		require.Equal(t, "test", cfg.FlightFetcherConfig.URL)
		require.Equal(t, "test", cfg.FlightFetcherConfig.User)
		require.Equal(t, "test", cfg.FlightFetcherConfig.Pass)
		require.Equal(t, "VHHH", cfg.SchedulerConfig.Airports)
		require.True(t, cfg.LoggerConfig.JSON)
		require.Equal(t, "info", cfg.LoggerConfig.Level)
	})
}

func TestLoadConfig_MissingFile_ShouldError(t *testing.T) {
	t.Run("Missing Config File", func(t *testing.T) {
		// Temporarily rename the config file if it exists
		originalPath := "../../config.yaml"
		tempPath := "../../config.yaml.bak"

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
	})
}

func TestLoadConfig_EnvOverride_ShouldSucceed(t *testing.T) {
	t.Run("Override Config File", func(t *testing.T) {
		t.Setenv("FLIGHT_READER_LOGGER_LEVEL", "debug")
		t.Setenv("FLIGHT_READER_LOGGER_JSON", "false")

		cfg, err := config.LoadConfig()

		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, "9099", cfg.HTTPServerConfig.Port)
		require.Equal(t, 10, cfg.HTTPServerConfig.Timeout)
		require.Equal(t, 10, cfg.HTTPClientConfig.Timeout)
		require.Equal(t, "localhost:9098", cfg.GrpcClientConfig.Address)
		require.Equal(t, "test", cfg.FlightFetcherConfig.URL)
		require.Equal(t, "test", cfg.FlightFetcherConfig.User)
		require.Equal(t, "test", cfg.FlightFetcherConfig.Pass)
		require.Equal(t, "VHHH", cfg.SchedulerConfig.Airports)
		require.False(t, cfg.LoggerConfig.JSON)
		require.Equal(t, "debug", cfg.LoggerConfig.Level)
	})
}

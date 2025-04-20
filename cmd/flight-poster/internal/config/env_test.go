package config_test

import (
	"os"
	"testing"

	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_ValidConfigFile_ShouldSucceed(t *testing.T) {
	t.Run("Valid Config File", func(t *testing.T) {
		os.Setenv("GOTWI_API_KEY", "test")
		os.Setenv("GOTWI_API_KEY_SECRET", "test")
		os.Setenv("FLIGHT_POSTER_THREADS_ACCESS_TOKEN", "test")
		os.Setenv("FLIGHT_POSTER_TWITTER_ACCESS_TOKEN_KEY", "test")
		os.Setenv("FLIGHT_POSTER_TWITTER_ACCESS_TOKEN_SECRET", "test")

		cfg, err := config.LoadConfig()

		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, "9097", cfg.GrpcServerConfig.Port)
		require.Equal(t, 10, cfg.HTTPClientConfig.Timeout)
		require.Equal(t, "https://graph.threads.net", cfg.ThreadsClientConfig.URL)
		require.Equal(t, "test", cfg.ThreadsClientConfig.Token)
		require.Equal(t, "test", cfg.TwitterClientConfig.Key)
		require.Equal(t, "test", cfg.TwitterClientConfig.Secret)
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
		t.Setenv("FLIGHT_POSTER_LOGGER_LEVEL", "debug")
		t.Setenv("FLIGHT_POSTER_LOGGER_JSON", "false")

		cfg, err := config.LoadConfig()

		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, "9097", cfg.GrpcServerConfig.Port)
		require.Equal(t, 10, cfg.HTTPClientConfig.Timeout)
		require.Equal(t, "https://graph.threads.net", cfg.ThreadsClientConfig.URL)
		require.Equal(t, "test", cfg.ThreadsClientConfig.Token)
		require.Equal(t, "test", cfg.TwitterClientConfig.Key)
		require.Equal(t, "test", cfg.TwitterClientConfig.Secret)
		require.False(t, cfg.LoggerConfig.JSON)
		require.Equal(t, "debug", cfg.LoggerConfig.Level)
	})
}

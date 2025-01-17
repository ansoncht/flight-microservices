package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/config"
	"github.com/stretchr/testify/require"
)

func makeValidTestConfig() string {
	content := `
grpc-server:
  port: "50051"
grpc-client:
  address: "localhost:50052"
mongodb:
  uri: ""
  db: "flight_db"
logger:
  json: "true"
  level: "info"
`
	//nolint:dogsled
	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)

	path := filepath.Join(dir, "config.yaml")

	tmpFile, _ := os.Create(path)

	_, _ = tmpFile.Write([]byte(content))
	tmpFile.Close()

	return path
}

func TestLoadConfig_ValidConfigFile_ShouldSucceed(t *testing.T) {
	t.Run("Valid Config File", func(t *testing.T) {
		path := makeValidTestConfig()
		defer os.Remove(path)

		os.Setenv("FLIGHT_PROCESSOR_MONGODB_URI", "mongodb://localhost:27017")

		cfg, err := config.LoadConfig()

		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, "50051", cfg.GrpcServerConfig.Port)
		require.Equal(t, "localhost:50052", cfg.GrpcClientConfig.Address)
		require.Equal(t, "mongodb://localhost:27017", cfg.MongoClientConfig.URI)
		require.Equal(t, "flight_db", cfg.MongoClientConfig.DB)
		require.True(t, cfg.LoggerConfig.JSON)
		require.Equal(t, "info", cfg.LoggerConfig.Level)
	})
}

func TestLoadConfig_MissingFile_ShouldError(t *testing.T) {
	t.Run("Missing Config File", func(t *testing.T) {
		cfg, err := config.LoadConfig()
		require.Error(t, err)
		require.Nil(t, cfg)
	})
}

func TestLoadConfig_EnvOverride_ShouldSucceed(t *testing.T) {
	t.Run("Override Config File", func(t *testing.T) {
		path := makeValidTestConfig()
		defer os.Remove(path)

		t.Setenv("FLIGHT_PROCESSOR_GRPC_SERVER_PORT", "60051")
		t.Setenv("FLIGHT_PROCESSOR_GRPC_CLIENT_ADDRESS", "localhost:60052")
		t.Setenv("FLIGHT_PROCESSOR_MONGODB_URI", "mongodb://localhost:37017")
		t.Setenv("FLIGHT_PROCESSOR_MONGODB_DB", "test_db")
		t.Setenv("FLIGHT_PROCESSOR_LOGGER_LEVEL", "debug")
		t.Setenv("FLIGHT_PROCESSOR_LOGGER_JSON", "false")

		cfg, err := config.LoadConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg)
		require.Equal(t, "60051", cfg.GrpcServerConfig.Port)
		require.Equal(t, "localhost:60052", cfg.GrpcClientConfig.Address)
		require.Equal(t, "mongodb://localhost:37017", cfg.MongoClientConfig.URI)
		require.Equal(t, "test_db", cfg.MongoClientConfig.DB)
		require.False(t, cfg.LoggerConfig.JSON)
		require.Equal(t, "debug", cfg.LoggerConfig.Level)
	})
}

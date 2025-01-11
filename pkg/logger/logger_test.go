package logger_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ansoncht/flight-microservices/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_ValidFormat_ShouldSucceed(t *testing.T) {
	tests := []struct {
		name  string
		cfg   *logger.Config
		level slog.Level
	}{
		{
			name: "DEBUG level with JSON format",
			cfg: &logger.Config{
				JSON:  true,
				Level: "debug",
			},
			level: slog.LevelDebug,
		},
		{
			name: "DEBUG level with TEXT format",
			cfg: &logger.Config{
				JSON:  false,
				Level: "debug",
			},
			level: slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			actual, err := logger.NewLogger(tt.cfg)

			require.NoError(t, err)
			require.NotNil(t, actual)
			require.True(t, actual.Handler().Enabled(ctx, tt.level))
		})
	}
}

func TestNewLogger_ValidLevel_ShouldSucceed(t *testing.T) {
	tests := []struct {
		name  string
		cfg   *logger.Config
		level slog.Level
	}{
		{
			name: "DEBUG level with JSON format",
			cfg: &logger.Config{
				JSON:  true,
				Level: "debug",
			},
			level: slog.LevelDebug,
		},
		{
			name: "INFO level with JSON format",
			cfg: &logger.Config{
				JSON:  true,
				Level: "info",
			},
			level: slog.LevelInfo,
		},
		{
			name: "WARN level with JSON format",
			cfg: &logger.Config{
				JSON:  true,
				Level: "warn",
			},
			level: slog.LevelWarn,
		},
		{
			name: "ERROR level with JSON format",
			cfg: &logger.Config{
				JSON:  true,
				Level: "error",
			},
			level: slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			actual, err := logger.NewLogger(tt.cfg)

			require.NoError(t, err)
			require.NotNil(t, actual)
			require.True(t, actual.Handler().Enabled(ctx, tt.level))
		})
	}
}

func TestNewLogger_InvalidLevel_ShouldError(t *testing.T) {
	tests := []struct {
		name  string
		cfg   *logger.Config
		level slog.Level
	}{
		{
			name: "DEBUG level with JSON format",
			cfg: &logger.Config{
				JSON:  true,
				Level: "",
			},
		},
		{
			name: "INFO level with JSON format",
			cfg: &logger.Config{
				JSON:  true,
				Level: "mylevel",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := logger.NewLogger(tt.cfg)

			require.ErrorContains(t, err, "invalid log level")
			require.Nil(t, actual)
		})
	}
}

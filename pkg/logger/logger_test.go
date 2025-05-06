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
		cfg   logger.Config
		level slog.Level
	}{
		{
			cfg: logger.Config{
				JSON:  true,
				Level: "debug",
			},
			level: slog.LevelDebug,
		},
		{
			cfg: logger.Config{
				JSON:  false,
				Level: "debug",
			},
			level: slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		ctx := context.Background()

		actual, err := logger.NewLogger(tt.cfg)

		require.NoError(t, err)
		require.NotNil(t, actual)
		require.True(t, actual.Handler().Enabled(ctx, tt.level))
	}
}

func TestNewLogger_ValidLevel_ShouldSucceed(t *testing.T) {
	tests := []struct {
		cfg   logger.Config
		level slog.Level
	}{
		{
			cfg: logger.Config{
				JSON:  true,
				Level: "debug",
			},
			level: slog.LevelDebug,
		},
		{
			cfg: logger.Config{
				JSON:  true,
				Level: "info",
			},
			level: slog.LevelInfo,
		},
		{
			cfg: logger.Config{
				JSON:  true,
				Level: "warn",
			},
			level: slog.LevelWarn,
		},
		{
			cfg: logger.Config{
				JSON:  true,
				Level: "error",
			},
			level: slog.LevelError,
		},
	}

	for _, tt := range tests {
		ctx := context.Background()

		actual, err := logger.NewLogger(tt.cfg)

		require.NoError(t, err)
		require.NotNil(t, actual)
		require.True(t, actual.Handler().Enabled(ctx, tt.level))
	}
}

func TestNewLogger_InvalidLevel_ShouldError(t *testing.T) {
	tests := []struct {
		cfg   logger.Config
		level slog.Level
	}{
		{
			cfg: logger.Config{
				JSON:  true,
				Level: "",
			},
		},
		{
			cfg: logger.Config{
				JSON:  true,
				Level: "mylevel",
			},
		},
	}

	for _, tt := range tests {
		actual, err := logger.NewLogger(tt.cfg)

		require.ErrorContains(t, err, "invalid log level")
		require.Equal(t, slog.Logger{}, actual)
	}
}

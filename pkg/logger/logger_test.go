package logger_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ansoncht/flight-microservices/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_ValidFormat_ShouldSucceed(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		cfg   logger.Config
		level slog.Level
	}{
		{
			name: "JSON format",
			cfg: logger.Config{
				JSON:  true,
				Level: "debug",
			},
			level: slog.LevelDebug,
		},
		{
			name: "Text format",
			cfg: logger.Config{
				JSON:  false,
				Level: "debug",
			},
			level: slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := logger.NewLogger(tt.cfg)
			require.NoError(t, err)
			require.NotNil(t, logger)
			require.True(t, logger.Handler().Enabled(ctx, tt.level))
		})
	}
}

func TestNewLogger_ValidLevel_ShouldSucceed(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		cfg   logger.Config
		level slog.Level
	}{
		{
			name: "Debug level",
			cfg: logger.Config{
				JSON:  true,
				Level: "debug",
			},
			level: slog.LevelDebug,
		},
		{
			name: "Info level",
			cfg: logger.Config{
				JSON:  true,
				Level: "info",
			},
			level: slog.LevelInfo,
		},
		{
			name: "Warn level",
			cfg: logger.Config{
				JSON:  true,
				Level: "warn",
			},
			level: slog.LevelWarn,
		},
		{
			name: "Error level",
			cfg: logger.Config{
				JSON:  true,
				Level: "error",
			},
			level: slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := logger.NewLogger(tt.cfg)
			require.NoError(t, err)
			require.NotNil(t, logger)
			require.True(t, logger.Handler().Enabled(ctx, tt.level))
		})
	}
}

func TestNewLogger_InvalidLevel_ShouldError(t *testing.T) {
	tests := []struct {
		name  string
		cfg   logger.Config
		level slog.Level
	}{
		{
			name: "Empty level",
			cfg: logger.Config{
				JSON:  true,
				Level: "",
			},
		},
		{
			name: "Invalid level",
			cfg: logger.Config{
				JSON:  true,
				Level: "mylevel",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := logger.NewLogger(tt.cfg)
			require.ErrorContains(t, err, "invalid log level")
			require.Equal(t, slog.Logger{}, logger)
		})
	}
}

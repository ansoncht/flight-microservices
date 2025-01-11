package logger

import (
	"fmt"
	"log/slog"
	"os"
)

// Config holds configuration settings for the logger.
type Config struct {
	// JSON specifies whether the logger output is in JSON format.
	JSON bool `mapstructure:"address"`
	// Level specifies the logging level (e.g., "debug", "info", "error").
	Level string `mapstructure:"address"`
}

// NewLogger creates a default logger based on the provided configuration.
func NewLogger(cfg *Config) (*slog.Logger, error) {
	slog.Debug("Initializing logger for the service")

	// Set log level based on the configuration.
	var level slog.Leveler
	addSource := false
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
		addSource = true
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid log level: %s", cfg.Level)
	}

	// Set source logging options.
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	}

	// Choose log format: text or JSON.
	handler = slog.NewTextHandler(os.Stdout, opts)
	if cfg.JSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler), nil
}

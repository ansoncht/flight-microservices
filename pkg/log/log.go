package log

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/caarlos0/env"
)

// LoggerConfig represents the configurations for the logger.
type LoggerConfig struct {
	JSON  bool   `env:"SERVICE_LOG_JSON"`  // Format of the logger
	Level string `env:"SERVICE_LOG_LEVEL"` // Level of the logger
}

// NewLogger creates and initializes the default logger.
func NewLogger() error {
	slog.Debug("Initializing logger for the service")

	loggerCfg, err := loadLoggerConfig()
	if err != nil {
		return fmt.Errorf("failed to get logger config: %w", err)
	}

	// Set slog level
	var level slog.Leveler
	addSource := false
	switch loggerCfg.Level {
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
		return fmt.Errorf("invalid log level: %s", loggerCfg.Level)
	}

	// Set slog source logging
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	}

	// Set slog format
	handler = slog.NewTextHandler(os.Stdout, opts)
	if loggerCfg.JSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))

	return nil
}

// loadLoggerConfig parses environment variables into a LoggerConfig struct.
func loadLoggerConfig() (*LoggerConfig, error) {
	var cfg LoggerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for logger config: %w", err)
	}

	return &cfg, nil
}

package log

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/caarlos0/env"
)

// LoggerConfig represents logger configurations.
type LoggerConfig struct {
	JSON  bool   `env:"LOGGER_JSON"`
	Level string `env:"LOGGER_LEVEL"`
}

// MakeLogger create and instantiate a default logger.
func MakeLogger() error {
	slog.Debug("Creating logger for the service")

	loggerCfg, err := parseConfig()
	if err != nil {
		return fmt.Errorf("failed to get logger config: %w", err)
	}

	// set slog level
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

	// set slog source logging
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	}

	// set slog format
	handler = slog.NewTextHandler(os.Stdout, opts)
	if loggerCfg.JSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))

	return nil
}

// parseConfig parses environment variables into a LoggerConfig struct.
func parseConfig() (*LoggerConfig, error) {
	var cfg LoggerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to get env for logger config: %w", err)
	}

	return &cfg, nil
}

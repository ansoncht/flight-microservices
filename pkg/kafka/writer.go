package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

// WriterConfig holds configuration settings for the Kafka writer.
type WriterConfig struct {
	// Address specifies the Kafka broker address.
	Address string `mapstructure:"address"`
	// Topic specifies the Kafka topic to write to.
	Topic string `mapstructure:"topic"`
}

// MessageWriter defines the interface for writing messages to a message queue.
type MessageWriter interface {
	// WriteMessage writes a message to the message queue.
	WriteMessage(ctx context.Context, key []byte, value []byte) error
}

// Writer holds the Kafka writer instance.
// It implements the MessageWriter interface to provide methods for writing messages to Kafka.
type Writer struct {
	// KafkaWriter specifies the kafka writer instance.
	KafkaWriter *kafka.Writer
}

// NewKafkaWriter creates a new Writer instance based on the provided configuration.
func NewKafkaWriter(cfg WriterConfig) (*Writer, error) {
	slog.Info("Initializing Kafka writer for the service", "address", cfg.Address, "topic", cfg.Topic)

	if cfg.Address == "" {
		return nil, fmt.Errorf("kafka broker address is empty")
	}

	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka topic is empty")
	}

	return &Writer{
		KafkaWriter: &kafka.Writer{
			Addr:  kafka.TCP(cfg.Address),
			Topic: cfg.Topic,
		},
	}, nil
}

// Close closes the Kafka writer.
func (w *Writer) Close() error {
	slog.Debug("Closing Kafka writer")

	if w == nil {
		return nil
	}

	if err := w.KafkaWriter.Close(); err != nil {
		slog.Error("Failed to close Kafka writer", "error", err)
		return fmt.Errorf("failed to close Kafka writer: %w", err)
	}

	slog.Info("Kafka writer closed successfully")

	return nil
}

// WriteMessage writes a message to the Kafka topic.
func (w *Writer) WriteMessage(ctx context.Context, key []byte, value []byte) error {
	slog.Debug("Writing message to Kafka topic")

	if w == nil {
		return fmt.Errorf("kafka writer is nil")
	}

	if len(key) == 0 {
		return fmt.Errorf("message key is nil or empty")
	}

	if len(value) == 0 {
		return fmt.Errorf("message value is nil or empty")
	}

	msg := kafka.Message{
		Key:   key,
		Value: value,
	}

	if err := w.KafkaWriter.WriteMessages(ctx, msg); err != nil {
		slog.Error("Failed to write message to Kafka topic", "error", err)
		return fmt.Errorf("failed to write message to Kafka topic: %w", err)
	}

	slog.Info("Message written to Kafka topic successfully", "key", string(key), "message", string(value))

	return nil
}

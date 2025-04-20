package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

// ReaderConfig holds configuration settings for the Kafka reader.
type ReaderConfig struct {
	// Address specifies the Kafka broker address.
	Address string `mapstructure:"address"`
	// Topic specifies the Kafka topic to read from.
	Topic string `mapstructure:"topic"`
	// GroupID specifies the consumer group ID.
	GroupID string `mapstructure:"group_id"`
}

// Reader holds the Kafka reader instance.
type Reader struct {
	// KafkaReader specifies the kafka reader instance.
	KafkaReader *kafka.Reader
}

// NewKafkaReader creates a new Reader instance based on the provided configuration.
func NewKafkaReader(cfg ReaderConfig) (*Reader, error) {
	slog.Info(
		"Initializing Kafka reader for the service",
		"address", cfg.Address,
		"topic", cfg.Topic,
		"group_id", cfg.GroupID,
	)

	// Validate the configuration
	if cfg.Address == "" {
		return nil, fmt.Errorf("kafka broker address is empty")
	}

	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka topic is empty")
	}

	if cfg.GroupID == "" {
		return nil, fmt.Errorf("kafka group ID is empty")
	}

	return &Reader{
		KafkaReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{cfg.Address},
			Topic:   cfg.Topic,
			GroupID: cfg.GroupID,
		}),
	}, nil
}

// Close closes the Kafka reader.
func (r *Reader) Close() error {
	slog.Debug("Closing Kafka reader")

	if r == nil {
		return nil
	}

	if err := r.KafkaReader.Close(); err != nil {
		slog.Error("Failed to close Kafka reader", "error", err)
		return fmt.Errorf("failed to close Kafka reader: %w", err)
	}

	slog.Info("Kafka reader closed successfully")

	return nil
}

// ReadMessages reads messages from the Kafka topic and sends them to the provided channel.
func (r *Reader) ReadMessages(ctx context.Context, msgChan chan<- kafka.Message) error {
	slog.Debug("Reading message from Kafka topic")

	if r == nil {
		return fmt.Errorf("kafka reader is nil")
	}

	defer close(msgChan)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping Kafka message reader due to context cancellation")
			return fmt.Errorf("context canceled while reading messages: %w", ctx.Err())
		default:
			// Read a message from Kafka
			message, err := r.KafkaReader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					slog.Info("Kafka reader reached EOF, stopping reader")
					return nil
				}
				if errors.Is(err, context.Canceled) {
					slog.Info("Kafka reader stopped due to context cancellation")
					return nil
				}
				slog.Error("Failed to read message from Kafka", "error", err)
				return fmt.Errorf("failed to read message from Kafka: %w", err)
			}

			// Send the message to the channel
			select {
			case msgChan <- message:
				slog.Debug(
					"Message sent to channel",
					"topic", message.Topic,
					"partition", message.Partition,
					"offset", message.Offset,
					"key", string(message.Key),
					"value", string(message.Value),
				)
			case <-ctx.Done():
				slog.Info("Stopping Kafka message reader due to context cancellation")
				return fmt.Errorf("context canceled while reading messages: %w", ctx.Err())
			}
		}
	}
}

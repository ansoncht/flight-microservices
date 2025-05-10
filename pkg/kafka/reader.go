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

// MessageReader defines the interface for reading messages from a message queue.
type MessageReader interface {
	// ReadMessages reads messages from the message queue.
	ReadMessages(ctx context.Context, msgChan chan<- kafka.Message) error
	// Close closes the message queue reader.
	Close() error
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
		}),
	}, nil
}

// Close closes the Kafka reader.
func (r *Reader) Close() error {
	if r == nil {
		return nil
	}

	if err := r.KafkaReader.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka reader: %w", err)
	}

	return nil
}

// ReadMessages reads messages from the Kafka topic and sends them to the provided channel.
func (r *Reader) ReadMessages(ctx context.Context, msgChan chan<- kafka.Message) error {
	slog.Info("Reading message from Kafka topic")

	defer close(msgChan)

	if r == nil {
		return fmt.Errorf("kafka reader is nil")
	}

readingLoop:
	for {
		select {
		case <-ctx.Done():
			break readingLoop
		default:
			// Read a message from Kafka
			message, err := r.KafkaReader.ReadMessage(ctx)

			if err != nil {
				if errors.Is(err, io.EOF) {
					slog.Info("Kafka reader reached EOF")
					continue
				}
				if errors.Is(err, context.Canceled) {
					slog.Info("Context canceled during ReadMessage")
					break readingLoop
				}
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
				slog.Info("Context is done, stopping message reading")
				break readingLoop
			}
		}
	}

	if ctx.Err() != nil {
		return fmt.Errorf("context canceled while reading kafka messages: %w", ctx.Err())
	}

	return nil
}

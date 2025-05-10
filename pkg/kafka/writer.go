package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"
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
	// Close closes the message queue writer.
	Close()
}

// Writer holds the Kafka writer instance.
// It implements the MessageWriter interface to provide methods for writing messages to Kafka.
type Writer struct {
	// Client specifies the kafka reader instance.
	Client *kgo.Client
}

// NewKafkaWriter creates a new Writer instance with the provided configuration.
func NewKafkaWriter(cfg WriterConfig) (*Writer, error) {
	slog.Info("Initializing Kafka writer for the service", "address", cfg.Address, "topic", cfg.Topic)

	if cfg.Address == "" {
		return nil, fmt.Errorf("kafka broker address is empty")
	}

	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka topic is empty")
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers([]string{cfg.Address}...),
		kgo.DefaultProduceTopic(cfg.Topic),
	}
	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	return &Writer{
		Client: client,
	}, nil
}

// Close closes the Kafka writer.
func (w *Writer) Close() {
	w.Client.Close()
}

// WriteMessage writes a message to the Kafka topic.
func (w *Writer) WriteMessage(ctx context.Context, key []byte, value []byte) error {
	slog.Info("Writing message to Kafka topic", "key", string(key), "message", string(value))

	if w == nil {
		return fmt.Errorf("kafka writer is nil")
	}

	if len(key) == 0 {
		return fmt.Errorf("message key is nil or empty")
	}

	if len(value) == 0 {
		return fmt.Errorf("message value is nil or empty")
	}

	record := &kgo.Record{
		Key:   key,
		Value: value,
	}

	errChan := make(chan error, 1)

	w.Client.Produce(ctx, record, func(_ *kgo.Record, err error) {
		if err != nil {
			errChan <- err
		} else {
			errChan <- nil
		}

		close(errChan)
	})

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("context canceled while producing message: %w", ctx.Err())
	}
}

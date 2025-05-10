package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	pollInterval = 5 * time.Second
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
	ReadMessages(ctx context.Context, msgChan chan<- kgo.Record) error
	// Close closes the message queue reader.
	Close()
}

// Reader holds the Kafka reader instance.
type Reader struct {
	// Client specifies the kafka client instance.
	Client *kgo.Client
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

	opts := []kgo.Opt{
		kgo.SeedBrokers([]string{cfg.Address}...),
		kgo.ConsumerGroup(cfg.GroupID),
		kgo.ConsumeTopics(cfg.Topic),
	}
	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	return &Reader{
		Client: client,
	}, nil
}

// Close closes the Kafka reader.
func (r *Reader) Close() {
	r.Client.Close()
}

// ReadMessages reads messages from the Kafka topic and sends them to the provided channel.
func (r *Reader) ReadMessages(ctx context.Context, msgChan chan<- kgo.Record) error {
	slog.Info("Reading message from Kafka topic")

	defer close(msgChan)

	if r == nil {
		return fmt.Errorf("kafka reader is nil")
	}

	// Poll every 5 seconds
	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()

readingLoop:
	for {
		select {
		case <-ctx.Done():
			break readingLoop
		case <-pollTicker.C:
			fetches := r.Client.PollFetches(ctx)

			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					if errors.Is(err.Err, context.Canceled) {
						break readingLoop
					}

					if err.Topic != "" || err.Partition != -1 || err.Err != nil {
						slog.Error("Failed to fetch message from Kafka", "errors", err)
					}
				}
			}

			// Process fetched messages
			fetches.EachPartition(func(p kgo.FetchTopicPartition) {
				for _, record := range p.Records {
					select {
					case <-ctx.Done():
						return
					case msgChan <- *record:
						r.Client.MarkCommitRecords(record)
					}
				}
			})
		}
	}

	if err := r.Client.CommitUncommittedOffsets(ctx); err != nil {
		return fmt.Errorf("failed to commit offsets: %w", err)
	}

	if ctx.Err() != nil {
		return fmt.Errorf("context canceled while fetching kafka messages: %w", ctx.Err())
	}

	return nil
}

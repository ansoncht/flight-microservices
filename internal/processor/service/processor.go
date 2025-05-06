package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	repo "github.com/ansoncht/flight-microservices/internal/processor/repository"
	msgQueue "github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/ansoncht/flight-microservices/pkg/model"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
)

// Processor holds dependencies for reading from Kafka, summarizing, storing, and publishing.
type Processor struct {
	// MessageWriter specifies the message writer to send messages to a message queue.
	MessageWriter msgQueue.MessageWriter
	// MessageReader specifies the message reader to read messages from a message queue.
	MessageReader msgQueue.MessageReader
	// summarizer specifies the summarizer to gather flights statistic.
	summarizer Summarizer
	// repository  specifies the repository to interact with the db collection.
	repository repo.SummaryRepository
}

// NewProcessor creates a new Processor instance based on the
// provided message writer, message reader, summarizer and repository.
func NewProcessor(
	messageWriter msgQueue.MessageWriter,
	messageReader msgQueue.MessageReader,
	summarizer Summarizer,
	repository repo.SummaryRepository,
) (*Processor, error) {
	if messageWriter == nil {
		return nil, fmt.Errorf("message writer is nil")
	}

	if messageReader == nil {
		return nil, fmt.Errorf("message reader is nil")
	}

	if summarizer == nil {
		return nil, fmt.Errorf("summarizer is nil")
	}

	if repository == nil {
		return nil, fmt.Errorf("repository is nil")
	}

	return &Processor{
		MessageWriter: messageWriter,
		MessageReader: messageReader,
		summarizer:    summarizer,
		repository:    repository,
	}, nil
}

func (p *Processor) Process(ctx context.Context) error {
	flights := make([]model.FlightRecord, 0)
	msgChan := make(chan kafka.Message)
	airport := ""

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return p.MessageReader.ReadMessages(gCtx, msgChan)
	})

processingLoop:
	for {
		select {
		case <-ctx.Done():
			break processingLoop
		case msg, ok := <-msgChan:
			if !ok {
				break processingLoop
			}

			key := string(msg.Key)

			switch key {
			case "start_of_stream":
				airport = string(msg.Value)
				slog.Info("Started processing stream for airport", "airport", airport)
			case "end_of_stream":
				endAirport := string(msg.Value)

				summary, err := p.summarizer.SummarizeFlights(flights, endAirport, airport)
				if err != nil {
					return fmt.Errorf("failed to summarize flights: %w", err)
				}

				objectID, err := p.repository.Insert(ctx, *summary)
				if err != nil {
					return fmt.Errorf("failed to insert summary: %w", err)
				}

				if err := p.MessageWriter.WriteMessage(ctx, []byte("summary_id"), []byte(objectID)); err != nil {
					return fmt.Errorf("failed to publish summary ObjectID: %w", err)
				}

				slog.Info("Published summary", "objectID", objectID)

				flights = flights[:0]
			default:
				flight, err := p.decodeMessage(msg.Value)
				if err != nil {
					slog.Warn("Failed to decode flight record", "key", key, "error", err)
					continue
				}

				flights = append(flights, *flight)
			}
		}
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error while reading messages: %w", err)
	}

	if ctx.Err() != nil {
		return fmt.Errorf("context canceled while processing messages: %w", ctx.Err())
	}

	return nil
}

// decodeMessage decodes the Kafka message to a FlightRecord.
func (p *Processor) decodeMessage(msg []byte) (*model.FlightRecord, error) {
	var flight model.FlightRecord
	if err := json.Unmarshal(msg, &flight); err != nil {
		return nil, fmt.Errorf("failed to parse flight record: %w", err)
	}

	return &flight, nil
}

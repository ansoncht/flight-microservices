package summarizer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/db"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/model"
)

type Summarizer struct {
	mongoDB      *db.MongoDB
	flightCounts map[string]int
}

func NewSummarizer(mongoDB *db.MongoDB) *Summarizer {
	return &Summarizer{
		mongoDB:      mongoDB,
		flightCounts: make(map[string]int),
	}
}

func (s *Summarizer) AddFlight(destination string) {
	slog.Debug("Adding flight to summary", "destination", destination)

	s.flightCounts[destination]++
}

func (s *Summarizer) StoreSummary(ctx context.Context, date string) error {
	summary := model.FlightSummary{
		Summary: s.flightCounts,
	}
	if err := s.mongoDB.InsertSummary(ctx, summary, date); err != nil {
		return fmt.Errorf("failed to insert summary: %w", err)
	}

	return nil
}

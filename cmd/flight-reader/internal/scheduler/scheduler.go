package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/client"
)

const hoursInADay = 24

type Scheduler struct {
	FlightFetcher client.FlightFetcher
}

// StartScheduler starts the scheduler to fetch flights every night at 2 AM.
func (s *Scheduler) ScheduleDailyFetch(ctx context.Context) error {
	// Calculate the duration until the next 2 AM
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if now.After(nextRun) {
		nextRun = nextRun.Add(hoursInADay * time.Hour)
	}

	durationUntilNextRun := nextRun.Sub(now)
	time.Sleep(durationUntilNextRun)

	if err := s.FlightFetcher.FetchFlightsFromAPI(ctx); err != nil {
		return fmt.Errorf("failed to fetch flights: %w", err)
	}

	ticker := time.NewTicker(hoursInADay * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.FlightFetcher.FetchFlightsFromAPI(ctx); err != nil {
				return fmt.Errorf("failed to fetch flights: %w", err)
			}
		case <-ctx.Done():
			slog.Info("Stopping scheduler due to context cancellation")
			return nil
		}
	}
}

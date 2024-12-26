package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/client"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/fetcher"
)

const (
	targetHour = 2
	hoursInDay = 24
)

type Scheduler struct {
	airports   []string
	grpcClient *client.GrpcClient
	fetchers   []fetcher.Fetcher
}

func NewScheduler(
	grpcClient *client.GrpcClient,
	fetchers []fetcher.Fetcher,
) (*Scheduler, error) {
	slog.Info("Creating scheduler for the service")

	cfg, err := config.LoadSchedulerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load scheduler config: %w", err)
	}

	if cfg.Airports == "" {
		return nil, fmt.Errorf("empty airports list")
	}

	// Split the airports string by commas
	airports := strings.Split(cfg.Airports, ",")
	for i, s := range airports {
		airports[i] = strings.TrimSpace(s)
	}

	return &Scheduler{
		airports:   airports,
		grpcClient: grpcClient,
		fetchers:   fetchers,
	}, nil
}

// ScheduleJob runs the job daily at 2 AM.
func (s *Scheduler) ScheduleJob(ctx context.Context) error {
	slog.Info("Starting Scheduler")

	for {
		// Calculate the next 2 AM time
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), targetHour, 0, 0, 0, now.Location())
		if now.After(nextRun) {
			nextRun = nextRun.Add(hoursInDay * time.Hour)
		}

		durationUntilNextRun := time.Until(nextRun)

		slog.Info("Next run scheduled", "time", nextRun)

		// Wait until the next run or context cancellation
		timer := time.NewTimer(durationUntilNextRun)

		select {
		case <-timer.C:
			for _, airport := range s.airports {
				if err := fetcher.ProcessFlights(ctx, s.grpcClient, s.fetchers, airport); err != nil {
					slog.Error("Failed to fetch flights", "error", err)
					return fmt.Errorf("failed to fetch flights: %w", err)
				}
			}

		case <-ctx.Done():
			slog.Info("Stopping scheduler due to context cancellation")
			timer.Stop()
			return nil
		}
	}
}

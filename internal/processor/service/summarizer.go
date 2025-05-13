package service

import (
	"fmt"
	"time"

	"github.com/ansoncht/flight-microservices/internal/processor/config"
	msg "github.com/ansoncht/flight-microservices/pkg/model"
)

const (
	format = "2006-01-02"
)

// Summarizer defines the interface for summarizing flight data.
type Summarizer interface {
	// SummarizeFlights summarizes flight data for a given date and airport.
	SummarizeFlights(records []msg.FlightRecord, date string, airport string) (*msg.DailyFlightSummary, error)
}

// FlightSummarizer implements the Summarizer interface.
type FlightSummarizer struct {
	// topN specifies the number of top airlines and destinations to return.
	topN int
}

// NewSummarizer creates a new Summarizer instance.
func NewSummarizer(cfg config.SummarizerConfig) (*FlightSummarizer, error) {
	if cfg.TopN <= 0 {
		return nil, fmt.Errorf("topN is invalid: %d", cfg.TopN)
	}

	return &FlightSummarizer{
		topN: cfg.TopN,
	}, nil
}

func (f *FlightSummarizer) SummarizeFlights(
	records []msg.FlightRecord,
	date string,
	airport string,
) (*msg.DailyFlightSummary, error) {
	dt, err := parseDate(date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date for transaction: %w", err)
	}

	airlineCounts := make(map[string]int)
	destCounts := make(map[string]int)
	totalFlights := 0

	for _, flight := range records {
		airlineCounts[flight.Airline]++
		destCounts[flight.Destination]++
		totalFlights++
	}

	// Get top n destinations and airlines
	topDestinations := topNKeysByValue(destCounts, f.topN)
	topAirlines := topNKeysByValue(airlineCounts, f.topN)

	return &msg.DailyFlightSummary{
		Date:              msg.ToMongoDateTime(dt),
		Airport:           airport,
		TotalFlights:      totalFlights,
		AirlineCounts:     airlineCounts,
		DestinationCounts: destCounts,
		TopDestinations:   topDestinations,
		TopAirlines:       topAirlines,
	}, nil
}

// topNKeysByValue returns the top N keys from a map[string]int by descending value.
func topNKeysByValue(m map[string]int, n int) []string {
	// get the maximum frequency in the map.
	var maxFreq int
	for _, val := range m {
		if val > maxFreq {
			maxFreq = val
		}
	}

	buckets := make([][]string, maxFreq+1)
	for key, val := range m {
		buckets[val] = append(buckets[val], key)
	}

	result := make([]string, 0, n)
	for i := maxFreq; i >= 0; i-- {
		for _, value := range buckets[i] {
			result = append(result, value)
			if len(result) == n {
				break
			}
		}
	}

	return result
}

// parseDate parses the date string into a time.Time object.
func parseDate(date string) (time.Time, error) {
	dt, err := time.Parse(format, date)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date: %w", err)
	}

	return dt, nil
}

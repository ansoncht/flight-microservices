package service_test

import (
	"testing"
	"time"

	"github.com/ansoncht/flight-microservices/internal/processor/config"
	"github.com/ansoncht/flight-microservices/internal/processor/service"
	msg "github.com/ansoncht/flight-microservices/pkg/model"
	"github.com/stretchr/testify/require"
)

func TestNewSummarizer_ValidConfig_ShouldSucceed(t *testing.T) {
	cfg := config.SummarizerConfig{
		TopN: 5,
	}

	summarizer, err := service.NewSummarizer(cfg)
	require.NoError(t, err)
	require.NotNil(t, summarizer)
}

func TestNewSummarizer_InValidConfig_ShouldError(t *testing.T) {
	cfg := config.SummarizerConfig{
		TopN: -1,
	}

	summarizer, err := service.NewSummarizer(cfg)
	require.ErrorContains(t, err, "topN is invalid")
	require.Nil(t, summarizer)
}

func TestSummarizeFlights_ValidAndEmptyData_ShouldSucceed(t *testing.T) {
	testCases := []struct {
		name            string
		flights         []msg.FlightRecord
		expectedSummary *msg.DailyFlightSummary
	}{
		{
			name: "Valid Data",
			flights: []msg.FlightRecord{
				{Airline: "United", FlightNumber: "101", Origin: "JFK", Destination: "LAX"},
				{Airline: "United", FlightNumber: "102", Origin: "LAX", Destination: "SFO"},
				{Airline: "Delta", FlightNumber: "201", Origin: "ATL", Destination: "LAX"},
				{Airline: "United", FlightNumber: "103", Origin: "ORD", Destination: "LAX"},
				{Airline: "American", FlightNumber: "301", Origin: "DFW", Destination: "LAX"},
				{Airline: "Delta", FlightNumber: "202", Origin: "JFK", Destination: "SFO"},
				{Airline: "United", FlightNumber: "104", Origin: "DEN", Destination: "LAX"},
				{Airline: "JetBlue", FlightNumber: "401", Origin: "BOS", Destination: "FLL"},
				{Airline: "United", FlightNumber: "105", Origin: "SFO", Destination: "JFK"},
				{Airline: "American", FlightNumber: "302", Origin: "LAX", Destination: "DFW"},
				{Airline: "United", FlightNumber: "106", Origin: "SAN", Destination: "SFO"},
			},
			expectedSummary: &msg.DailyFlightSummary{
				Date:              msg.ToMongoDateTime(time.Date(2025, 5, 7, 0, 0, 0, 0, time.UTC)),
				Airport:           "SFO",
				TotalFlights:      11,
				AirlineCounts:     map[string]int{"United": 6, "Delta": 2, "American": 2, "JetBlue": 1},
				DestinationCounts: map[string]int{"LAX": 5, "SFO": 3, "JFK": 1, "DFW": 1, "FLL": 1},
				TopDestinations:   []string{"LAX", "SFO", "JFK", "DFW", "FLL"},
				TopAirlines:       []string{"United", "Delta", "American", "JetBlue"},
			},
		},
		{
			name:    "Empty Flights",
			flights: []msg.FlightRecord{},
			expectedSummary: &msg.DailyFlightSummary{
				Date:              msg.ToMongoDateTime(time.Date(2025, 5, 7, 0, 0, 0, 0, time.UTC)),
				Airport:           "SFO",
				TotalFlights:      0,
				AirlineCounts:     map[string]int{},
				DestinationCounts: map[string]int{},
				TopDestinations:   []string{},
				TopAirlines:       []string{},
			},
		},
	}

	cfg := config.SummarizerConfig{
		TopN: 5,
	}
	summarizer, err := service.NewSummarizer(cfg)
	require.NoError(t, err)
	require.NotNil(t, summarizer)

	date := "2025-05-07"
	airport := "SFO"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			summary, err := summarizer.SummarizeFlights(tc.flights, date, airport)
			require.NoError(t, err)
			require.NotNil(t, summary)
			require.Equal(t, tc.expectedSummary.Date, summary.Date)
			require.Equal(t, tc.expectedSummary.Airport, summary.Airport)
			require.Equal(t, tc.expectedSummary.TotalFlights, summary.TotalFlights)
			require.Equal(t, tc.expectedSummary.AirlineCounts, summary.AirlineCounts)
			require.Equal(t, tc.expectedSummary.DestinationCounts, summary.DestinationCounts)
			require.ElementsMatch(t, tc.expectedSummary.TopDestinations, summary.TopDestinations)
			require.ElementsMatch(t, tc.expectedSummary.TopAirlines, summary.TopAirlines)
		})
	}
}

func TestSummarizeFlights_InvalidData_ShouldError(t *testing.T) {
	testCases := []struct {
		name        string
		date        string
		airport     string
		flights     []msg.FlightRecord
		expectedErr string
	}{
		{
			name:        "Invalid Date Format",
			date:        "invalid-date",
			airport:     "SFO",
			flights:     []msg.FlightRecord{{Airline: "United", FlightNumber: "101", Origin: "JFK", Destination: "LAX"}},
			expectedErr: "failed to parse date for transaction",
		},
	}

	cfg := config.SummarizerConfig{
		TopN: 5,
	}
	summarizer, err := service.NewSummarizer(cfg)
	require.NoError(t, err)
	require.NotNil(t, summarizer)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			summary, err := summarizer.SummarizeFlights(tc.flights, tc.date, tc.airport)
			require.ErrorContains(t, err, tc.expectedErr)
			require.Nil(t, summary)
		})
	}
}

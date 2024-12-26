package fetcher

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/client"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/model"
	"golang.org/x/sync/errgroup"
)

// Fetcher interface for fetching data.
type Fetcher interface {
	Fetch(ctx context.Context, param ...string) (interface{}, error)
}

// ProcessFlights fetches flight data for a specified airport, processes
// the data to retrieve route information, and sends the flight details
// to a gRPC client for further processing.
func ProcessFlights(
	ctx context.Context,
	client *client.GrpcClient,
	fetchers []Fetcher,
	airport string,
) error {
	// Get previous day in Unix timestamp
	begin, end := getPreviousDayTime()

	// Fetch flights from the flightFetcher
	flightEntries, err := fetchFlights(ctx, fetchers[0], airport, begin, end)
	if err != nil {
		return err
	}

	// Start gRPC stream for the current batch of flights
	if err := client.StartStream(ctx); err != nil {
		return fmt.Errorf("failed to start gRPC stream: %w", err)
	}

	// Use errgroup with shared context to process routes concurrently
	g, ctx := errgroup.WithContext(ctx)

	// For each flight entry, process its route concurrently
	for _, entry := range flightEntries {
		g.Go(func() error {
			if err := processFlightRoutes(ctx, entry, fetchers[1], client); err != nil {
				return err
			}
			return nil
		})
	}

	// Wait for all goroutines to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("processing encountered an error: %w", err)
	}

	// Close gRPC stream to avoid long idle connection
	if err := client.CloseStream(); err != nil {
		return fmt.Errorf("failed to close gRPC stream: %w", err)
	}

	return nil
}

// fetchFlights fetches flight data from the flightFetcher for a given airport
// and time range (previous day).
func fetchFlights(
	ctx context.Context,
	flightFetcher Fetcher,
	airport string,
	begin string,
	end string,
) ([]*model.FlightResponse, error) {
	// Make API call to fetch flights for the specified airport
	flightResp, err := flightFetcher.Fetch(ctx, airport, begin, end)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch flights: %w", err)
	}

	// Assert that the fetched data is of type []*model.Flight
	entries, ok := flightResp.([]*model.FlightResponse)
	if !ok {
		slog.Error(
			"Type assertion failed",
			"expected", "[]*model.FlightResponse",
			"got", fmt.Sprintf("%T", flightResp),
		)
		return nil, fmt.Errorf("fetched data is not of type []*model.FlightResponse")
	}

	return entries, nil
}

// processFlightRoutes processes flight routes for a single flight entry
// by fetching route information and sending the flight data via gRPC.
func processFlightRoutes(
	ctx context.Context,
	entry *model.FlightResponse,
	routeFetcher Fetcher,
	client *client.GrpcClient,
) error {
	// Fetch route data for the flight using the routeFetcher
	routeResp, err := routeFetcher.Fetch(ctx, strings.TrimSpace(entry.CallSign))
	if err != nil {
		slog.Warn(
			"Cannot identify call sign, skipping with trace",
			"callsign", strings.TrimSpace(entry.CallSign),
			"error", err,
		)
		return nil
	}

	// Type assertion to ensure the fetched route data is of the expected type
	route, ok := routeResp.(*model.RouteResponse)
	if !ok {
		return fmt.Errorf("fetched data is not of type model.RouteResponse")
	}

	// Create a Flight object from the route data
	flight := &model.Flight{
		FlightNumber: route.Response.FlightRoute.CallSignIATA,
		Airline:      route.Response.FlightRoute.Airline.IATA,
		Origin:       route.Response.FlightRoute.Origin.IATACode,
		Destination:  route.Response.FlightRoute.Destination.IATACode,
		FirstSeen:    entry.FirstSeen,
		LastSeen:     entry.LastSeen,
	}

	// Log flight details for tracking
	slog.Info(
		"Flight processed",
		"flight_number", flight.FlightNumber,
		"origin", flight.Origin,
		"destination", flight.Destination,
	)

	// Send flight data to the processor via gRPC
	if err := client.SendFlight(flight); err != nil {
		return fmt.Errorf("failed to send flight to processor: %w", err)
	}

	return nil
}

// getPreviousDayTime calculates the start and end Unix timestamps for the previous day.
func getPreviousDayTime() (string, string) {
	now := time.Now()

	// Mark the start of yesterday (12:00:00 AM)
	startOfYesterday := time.Date(now.Year(), now.Month(), now.Day()-2, 0, 0, 0, 0, now.Location())

	// Mark the end of yesterday (11:59:59 PM)
	endOfYesterday := time.Date(now.Year(), now.Month(), now.Day()-2, 23, 59, 59, 0, now.Location())

	// Convert to Unix epoch timestamps
	startEpoch := startOfYesterday.Unix()
	endEpoch := endOfYesterday.Unix()

	return fmt.Sprintf("%d", startEpoch), fmt.Sprintf("%d", endEpoch)
}

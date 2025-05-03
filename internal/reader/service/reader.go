package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ansoncht/flight-microservices/internal/reader/client"
	"github.com/ansoncht/flight-microservices/internal/reader/model"
	"github.com/ansoncht/flight-microservices/pkg/kafka"
	msg "github.com/ansoncht/flight-microservices/pkg/model"
	"golang.org/x/sync/errgroup"
)

type Reader struct {
	// flightsClient specifies the shared HTTP client to submit requests to the external flight API.
	flightsClient client.FlightsClient
	// routesClient specifies the shared HTTP client to submit requests to the external route API.
	routeClient client.RoutesClient
	// messageWriter specifies the message writer to send messages to a message queue.
	messageWriter kafka.MessageWriter
}

// NewReader creates a new Reader instance based on the provided api clients and meesage writer.
func NewReader(
	flightClient client.FlightsClient,
	routeClient client.RoutesClient,
	messageWriter kafka.MessageWriter,
) (*Reader, error) {
	if flightClient == nil {
		return nil, fmt.Errorf("flight client is nil")
	}

	if routeClient == nil {
		return nil, fmt.Errorf("route client is nil")
	}

	if messageWriter == nil {
		return nil, fmt.Errorf("message writer is nil")
	}

	return &Reader{
		flightsClient: flightClient,
		routeClient:   routeClient,
		messageWriter: messageWriter,
	}, nil
}

func (r *Reader) HTTPHandler(w http.ResponseWriter, req *http.Request) {
	airport := req.URL.Query().Get("airport")
	if airport == "" {
		http.Error(w, "missing airport parameter", http.StatusBadRequest)
		return
	}

	err := r.processFlights(req.Context(), airport)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to process flights: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte("Flights processed successfully")); err != nil {
		slog.Error("Failed to write HTTP response", "error", err)
	}
}

// processFlights fetches flight data for a specified airport, processes
// the data to retrieve route information, and sends the flight details
// to a kafka topic.
func (r *Reader) processFlights(
	ctx context.Context,
	airport string,
) error {
	// Get previous day in Unix timestamp
	begin, end := getPreviousDayTime()

	flights, err := r.flightsClient.FetchFlights(ctx, airport, begin, end)
	if err != nil {
		slog.Error("Failed to process flights", "error", err)
		return fmt.Errorf("failed to process flights: %w", err)
	}

	if err := r.processRoute(ctx, flights); err != nil {
		slog.Error("Failed to process routes", "error", err)
		return fmt.Errorf("failed to process routes: %w", err)
	}

	return nil
}

func (r *Reader) processRoute(ctx context.Context, flights []model.Flight) error {
	// Use errgroup with shared context to process routes concurrently
	g, gCtx := errgroup.WithContext(ctx)

	// For each flight entry, process its route concurrently
	for _, flight := range flights {
		if flight.Origin != "" && flight.Destination != "" && flight.Origin != flight.Destination {
			g.Go(func() error {
				callsign := strings.TrimSpace(flight.Callsign)
				if callsign == "" {
					slog.Warn("Empty callsign, skipping flight", "flight", flight)
					return nil
				}

				route, err := r.routeClient.FetchRoute(gCtx, callsign)
				if err != nil {
					slog.Warn("Failed to fetch route", "callsign", callsign, "error", err)
					// return fmt.Errorf("failed to fetch route for %s: %w", callsign, err)
					return nil
				}

				// Send the flight and route data to a message queue
				if err := r.sendFlightAndRouteMessage(gCtx, flight, *route); err != nil {
					slog.Warn("Failed to send flight and route message", "callsign", callsign, "error", err)
					// return fmt.Errorf("failed to send flight and route message for %s: %w", callsign, err)
					return nil
				}

				slog.Debug("Successfully sent flight and route message", "callsign", callsign)

				return nil
			})
		}
	}

	// Wait for all goroutines to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to process at least one route: %w", err)
	}

	return nil
}

func (r *Reader) sendFlightAndRouteMessage(
	ctx context.Context,
	flight model.Flight,
	route model.Route,
) error {
	record := &msg.FlightRecord{
		FlightNumber: route.Response.FlightRoute.CallSignIATA,
		Airline:      route.Response.FlightRoute.Airline.Name,
		Origin:       route.Response.FlightRoute.Origin.IATACode,
		Destination:  route.Response.FlightRoute.Destination.IATACode,
		FirstSeen:    flight.FirstSeen,
		LastSeen:     flight.LastSeen,
	}

	value, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal flight record: %w", err)
	}

	key := []byte(route.Response.FlightRoute.CallSignIATA)

	if err := r.messageWriter.WriteMessage(ctx, key, value); err != nil {
		slog.Error("Failed to write message to the message queue", "error", err)
		return fmt.Errorf("failed to write message to the message queue: %w", err)
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

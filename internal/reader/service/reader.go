package service

import (
	"context"
	"encoding/json"
	"errors"
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
	flightsClient client.Flight
	// routesClient specifies the shared HTTP client to submit requests to the external route API.
	routeClient client.Route
	// messageWriter specifies the message writer to send messages to a message queue.
	messageWriter kafka.MessageWriter
}

// NewReader creates a new Reader instance based on the provided api clients and message writer.
func NewReader(
	flightClient client.Flight,
	routeClient client.Route,
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

// Close closes the reader service.
func (r *Reader) Close() {
	r.messageWriter.Close()
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

	// Set the response header for JSON content type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode a success message as JSON
	response := map[string]string{"message": "flights processed successfully"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to write response", "error", err)
		http.Error(w, "failed to send response", http.StatusInternalServerError)
		return
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
	begin, end, date := getPreviousDayTime()

	flights, err := r.flightsClient.FetchFlights(ctx, airport, begin, end)
	if err != nil {
		return fmt.Errorf("failed to process flights: %w", err)
	}

	slog.Info("Fetched flights successfully", "airport", airport, "flights_count", len(flights))

	if err := r.processRoute(ctx, flights, airport, date); err != nil {
		return fmt.Errorf("failed to process routes: %w", err)
	}

	return nil
}

func (r *Reader) processRoute(ctx context.Context, flights []model.Flight, airport string, date string) error {
	if err := r.sendStreamControlMessage(ctx, "start_of_stream", airport); err != nil {
		return fmt.Errorf("failed to send start_of_stream message: %w", err)
	}

	// Use errgroup with shared context to process routes concurrently
	g, gCtx := errgroup.WithContext(ctx)

	// For each flight entry, process its route concurrently
	for _, f := range flights {
		flight := f
		if flight.Origin != "" && flight.Destination != "" && flight.Origin != flight.Destination {
			g.Go(func() error {
				callsign := strings.TrimSpace(flight.Callsign)
				if callsign == "" {
					slog.Warn("Empty callsign, skipping flight", "flight", flight)
					return nil
				}

				route, err := r.routeClient.FetchRoute(gCtx, callsign)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return fmt.Errorf("context canceled while processing route: %w", gCtx.Err())
					}

					slog.Warn("Failed to fetch route", "callsign", callsign, "error", err)
					return nil
				}

				// Send the flight and route data to a message queue
				if err := r.sendFlightAndRouteMessage(gCtx, flight, *route); err != nil {
					if errors.Is(err, context.Canceled) {
						return fmt.Errorf("context canceled while sending flight and route: %w", gCtx.Err())
					}

					slog.Warn("Failed to send flight and route message", "callsign", callsign, "error", err)
					return nil
				}

				return nil
			})
		}
	}

	// Wait for all goroutines to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to process at least one route: %w", err)
	}

	if err := r.sendStreamControlMessage(ctx, "end_of_stream", date); err != nil {
		return fmt.Errorf("failed to send end_of_stream message: %w", err)
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
		return fmt.Errorf("failed to write message to the message queue: %w", err)
	}

	return nil
}

func (r *Reader) sendStreamControlMessage(ctx context.Context, key, message string) error {
	if err := r.messageWriter.WriteMessage(ctx, []byte(key), []byte(message)); err != nil {
		return fmt.Errorf("failed to write %s message to the message queue: %w", key, err)
	}

	return nil
}

// getPreviousDayTime calculates the start and end Unix timestamps for the previous day.
func getPreviousDayTime() (string, string, string) {
	now := time.Now()

	yesterday := now.AddDate(0, 0, -2)

	// Mark the start of yesterday (12:00:00 AM)
	startOfYesterday := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())

	// Mark the end of yesterday (11:59:59 PM)
	endOfYesterday := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, now.Location())

	// Convert to Unix epoch timestamps
	startEpoch := startOfYesterday.Unix()
	endEpoch := endOfYesterday.Unix()

	date := yesterday.Format("2006-01-02")

	return fmt.Sprintf("%d", startEpoch), fmt.Sprintf("%d", endEpoch), date
}

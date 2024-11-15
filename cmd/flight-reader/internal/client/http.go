package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/model"
	pb "github.com/ansoncht/flight-microservices/proto/src/summarizer"
)

const (
	statusOK = http.StatusOK
)

type FlightFetcher interface {
	FetchFlightsFromAPI(context.Context) error
}

// HTTPClient represents http client and endpoint it is fetching from.
type HTTPClient struct {
	Client   *http.Client
	Endpoint string
	GRPC     pb.SummarizerClient
}

// FetchFlightsFromAPI fetch information from the predefined API.
func (c *HTTPClient) FetchFlightsFromAPI(ctx context.Context) error {
	slog.Info("fetching data from api", "url", c.Endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != statusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if resp.Body == nil {
		return fmt.Errorf("response body is nil")
	}

	flightData, err := c.decodeAPIResponse(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := c.printFlightData(ctx, flightData); err != nil {
		return fmt.Errorf("failed to send to processor: %w", err)
	}

	return nil
}

// decodeResponse decodes the JSON response body into flight data model.
func (c *HTTPClient) decodeAPIResponse(body io.ReadCloser) ([]model.FlightData, error) {
	// Decode the JSON response
	var flightData []model.FlightData
	if err := json.NewDecoder(body).Decode(&flightData); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return flightData, nil
}

// printFlightData prints decoded flight data in format.
func (c *HTTPClient) printFlightData(ctx context.Context, flightData []model.FlightData) error {
	if len(flightData) == 0 {
		slog.Warn("no flight data found in response")

		return nil
	}

	stream, err := c.GRPC.PullFlight(ctx)
	if err != nil {
		return fmt.Errorf("failed to create gRPC stream: %w", err)
	}

	data := flightData[0]
	for _, flightList := range data.List {
		var lastDestination, flightNo string

		fmt.Printf("  Time: %s, Status: %s, Terminal: %s, Gate: %s\n",
			flightList.Time, flightList.Status, flightList.Terminal, flightList.Gate)
		for _, flight := range flightList.Flight {
			fmt.Printf("    Flight No: %s, Airline: %s\n", flight.No, flight.Airline)
			flightNo = flight.No
			break
		}

		if len(flightList.Destination) > 0 {
			lastDestination = flightList.Destination[len(flightList.Destination)-1]
			fmt.Printf("    Last Destination: %s\n", lastDestination)
		}

		if err := stream.Send(&pb.PullFlightRequest{
			Flight:      flightNo,
			Origin:      "HKG",
			Destination: lastDestination,
		}); err != nil {
			return fmt.Errorf("send stream: %w", err)
		}
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to close stream: %w", err)
	}

	slog.Info("received response when closing stream", "total", strconv.Itoa(int(response.Transaction)))
	return nil
}

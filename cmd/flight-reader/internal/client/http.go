package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/model"
)

// HTTPClient represents http client and endpoint it is fetching from.
type HTTPClient struct {
	Client   *http.Client
	Endpoint string
}

// FetchFlightsFromAPI fetch information from the predefined API
func (c *HTTPClient) FetchFlightsFromAPI(ctx context.Context) error {
	slog.Info("fetching data from api", "url", c.Endpoint)

	resp, err := c.Client.Get(c.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch from API: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the JSON response
	var flightDataList []model.FlightData
	if err := json.NewDecoder(resp.Body).Decode(&flightDataList); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}

	if len(flightDataList) == 0 {
		slog.Warn("no flight data found in response")
	}

	flightData := flightDataList[0]

	// Print the decoded response
	for _, flightList := range flightData.List {
		fmt.Printf("  Time: %s, Status: %s, Terminal: %s, Gate: %s\n",
			flightList.Time, flightList.Status, flightList.Terminal, flightList.Gate)
		for _, flight := range flightList.Flight {
			fmt.Printf("    Flight No: %s, Airline: %s\n", flight.No, flight.Airline)
		}
		// Print the last destination if needed
		if len(flightList.Destination) > 0 {
			lastDestination := flightList.Destination[len(flightList.Destination)-1]
			fmt.Printf("    Last Destination: %s\n", lastDestination)
		}
	}

	return nil
}

package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/model"
)

// FlightFetcher is responsible for fetching flight list from an external API.
// FlightFetcher implements the Fetcher interface.
type FlightFetcher struct {
	client  *http.Client // Shared HTTP client to submit requests
	BaseURL string       // Base URL for the flight API
	User    string       // Username for the flight API
	Pass    string       // Password for the flight API
}

// NewFlightFetcher creates a new instance of FlightFetcher with the provided BaseFetcher client.
func NewFlightFetcher(client *http.Client) (*FlightFetcher, error) {
	slog.Info("Creating Flight Fetcher for the service")

	cfg, err := config.LoadFlightFetcherConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load flight fetcher config: %w", err)
	}

	return &FlightFetcher{
		client:  client,
		BaseURL: cfg.URL,
		User:    cfg.User,
		Pass:    cfg.Pass,
	}, nil
}

// Fetch retrieves a list of flights from the API.
func (f *FlightFetcher) Fetch(ctx context.Context, param ...string) (interface{}, error) {
	// Parse the base URL
	endpoint, err := url.Parse(f.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	// Add path segments
	endpoint = endpoint.JoinPath("api", "flights", "departure")

	// Add and set query parameters
	query := endpoint.Query()
	query.Add("airport", param[0])
	query.Add("begin", param[1])
	query.Add("end", param[2])
	endpoint.RawQuery = query.Encode()

	// Create a HTTP GET request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for flight: %w", err)
	}

	// Set basic authentication
	req.SetBasicAuth(f.User, f.Pass)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch flight: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response body into flight data
	flights, err := f.decodeReponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return flights, nil
}

// decodeReponse decodes the JSON response body into a slice of FlightResponse models.
func (f *FlightFetcher) decodeReponse(body io.ReadCloser) ([]*model.FlightResponse, error) {
	var flights []*model.FlightResponse
	if err := json.NewDecoder(body).Decode(&flights); err != nil {
		return nil, fmt.Errorf("failed to parse flights: %w", err)
	}

	return flights, nil
}

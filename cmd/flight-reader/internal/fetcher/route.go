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

// RouteFetcher is responsible for fetching route from an external API.
// RouteFetcher implements the Fetcher interface.
type RouteFetcher struct {
	client  *http.Client // Shared HTTP client to submit requests
	BaseURL string       // Base URL for the route API
}

// NewRouteFetcher creates a new instance of RouteFetcher with the provided BaseFetcher client.
func NewRouteFetcher(client *http.Client) (*RouteFetcher, error) {
	slog.Info("Creating Route Fetcher for the service")

	cfg, err := config.LoadRouteFetcherConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load route fetcher config: %w", err)
	}

	return &RouteFetcher{
		client:  client,
		BaseURL: cfg.URL,
	}, nil
}

func (r *RouteFetcher) Fetch(ctx context.Context, param ...string) (interface{}, error) {
	// Parse the base URL
	endpoint, err := url.Parse(r.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	// Add path segmenta
	endpoint = endpoint.JoinPath("v0", "callsign", param[0])

	// Create a HTTP GET request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for route: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response body into route data
	route, err := r.decodeReponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return route, nil
}

// decodeReponse decodes the JSON response body into a RouteResponse model.
func (r *RouteFetcher) decodeReponse(body io.ReadCloser) (*model.RouteResponse, error) {
	var route *model.RouteResponse
	if err := json.NewDecoder(body).Decode(&route); err != nil {
		return nil, fmt.Errorf("failed to parse route: %w", err)
	}

	return route, nil
}

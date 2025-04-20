package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ansoncht/flight-microservices/internal/reader/config"
	"github.com/ansoncht/flight-microservices/internal/reader/model"
)

// RoutesClient defines the interface for fetching flight routes.
type RoutesClient interface {
	// FetchRoute retrieves the flight route for a given callsign from external API.
	FetchRoute(ctx context.Context, callsign string) (*model.Route, error)
}

// RouteAPIClient holds the configuration for fetching flight routes from an external API.
// It implements the RoutesClient interface to provide methods for fetching flight routes.
type RouteAPIClient struct {
	// client specifies the shared HTTP client to submit requests to the external route API.
	client *http.Client
	// BaseURL specifies the base URL for the external route API.
	BaseURL string
}

// NewRouteAPIClient creates a new outeAPIClient instance based on the provided configuration and HTTP client.
func NewRouteAPIClient(cfg config.RouteAPIClientConfig, client *http.Client) (*RouteAPIClient, error) {
	slog.Info("Initializing route API client", "url", cfg.URL)

	if client == nil {
		return nil, fmt.Errorf("http client is nil")
	}

	// Validate the configuration
	if cfg.URL == "" {
		return nil, fmt.Errorf("route api url is empty")
	}

	return &RouteAPIClient{
		client:  client,
		BaseURL: cfg.URL,
	}, nil
}

// FetchRoute retrieves the flight route from the external API based on the provided callsign.
func (c *RouteAPIClient) FetchRoute(ctx context.Context, callsign string) (*model.Route, error) {
	// Parse the base URL
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	// Add path segments
	endpoint = endpoint.JoinPath("v0", "callsign", callsign)

	// Create a HTTP GET request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for route: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response body into route data
	route, err := c.decodeReponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &route, nil
}

// decodeReponse decodes the JSON response body into a Route model.
func (c *RouteAPIClient) decodeReponse(body io.ReadCloser) (model.Route, error) {
	var route model.Route
	if err := json.NewDecoder(body).Decode(&route); err != nil {
		return model.Route{}, fmt.Errorf("failed to parse route: %w", err)
	}

	return route, nil
}

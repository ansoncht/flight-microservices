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

// FlightsClient defines the interface for fetching flight data.
type FlightsClient interface {
	// FetchFlights retrieves a list of flights from external API.
	FetchFlights(ctx context.Context, airportCode string, start string, end string) ([]model.Flight, error)
}

// FlightAPIClient holds the configuration for fetching flight data from an external API.
// It implements the FlightsClient interface to provide methods for fetching flight data.
type FlightAPIClient struct {
	// client specifies the shared HTTP client to submit requests to the external flight API.
	client *http.Client
	// BaseURL specifies the base URL for the external flight API.
	BaseURL string
	// User specifies the username for the external flight API.
	User string
	// Pass specifies the password for the external flight API.
	Pass string
}

// NewFlightAPIClient creates a new FlightAPIClient instance based on the provided configuration and HTTP client.
func NewFlightAPIClient(cfg config.FlightAPIClientConfig, client *http.Client) (*FlightAPIClient, error) {
	slog.Info("Initializing flight API client", "url", cfg.URL)

	if client == nil {
		return nil, fmt.Errorf("http client is nil")
	}

	// Validate the configuration
	if cfg.URL == "" {
		return nil, fmt.Errorf("flight api url is empty")
	}

	if cfg.User == "" {
		return nil, fmt.Errorf("flight api user is empty")
	}

	if cfg.Pass == "" {
		return nil, fmt.Errorf("flight api password is empty")
	}

	return &FlightAPIClient{
		client:  client,
		BaseURL: cfg.URL,
		User:    cfg.User,
		Pass:    cfg.Pass,
	}, nil
}

// FetchFlights retrieves a list of flights from the external API based on the provided airport code and time range.
func (c *FlightAPIClient) FetchFlights(
	ctx context.Context,
	airportCode string,
	start string,
	end string,
) ([]model.Flight, error) {
	// Parse the base URL
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	// Add path segments
	endpoint = endpoint.JoinPath("api", "flights", "departure")

	// Add and set query parameters
	query := endpoint.Query()
	query.Add("airport", airportCode)
	query.Add("begin", start)
	query.Add("end", end)
	endpoint.RawQuery = query.Encode()

	// Create a HTTP GET request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for flight: %w", err)
	}

	// Set basic authentication
	req.SetBasicAuth(c.User, c.Pass)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch flight: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response body into flight data
	flights, err := c.decodeReponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return flights, nil
}

// decodeReponse decodes the JSON response body into a slice of Flight models.
func (c *FlightAPIClient) decodeReponse(body io.ReadCloser) ([]model.Flight, error) {
	var flights []model.Flight
	if err := json.NewDecoder(body).Decode(&flights); err != nil {
		return nil, fmt.Errorf("failed to parse flights: %w", err)
	}

	return flights, nil
}

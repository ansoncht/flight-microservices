package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/model"
)

const (
	statusOK = http.StatusOK
)

type FlightFetcher interface {
	FetchFlights(ctx context.Context) ([]model.FlightData, error)
}

type AirportFetcher struct {
	// client  *http.Client
	// airport string
	// url     string
}

type HTTPClient struct {
	client *http.Client
	// fetchers []FlightFetcher
}

func NewHTTPClient() (*HTTPClient, error) {
	slog.Debug("Creating http client for the service")

	cfg, err := config.MakeHTTPClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get http client config: %w", err)
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}, nil
}

// FetchFlightsFromAPI fetch information from the predefined API and returns
// the flight data.
func (c *HTTPClient) FetchFlightsFromAPI(
	ctx context.Context,
	url string,
) ([]model.FlightData, error) {
	slog.Info("Fetching data from api", "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != statusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	flightData, err := c.decodeAPIResponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return flightData, nil
}

// decodeResponse decodes the JSON response body into flight data model.
func (c *HTTPClient) decodeAPIResponse(body io.ReadCloser) ([]model.FlightData, error) {
	var flightData []model.FlightData
	if err := json.NewDecoder(body).Decode(&flightData); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return flightData, nil
}

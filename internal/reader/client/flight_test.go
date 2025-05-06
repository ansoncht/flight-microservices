package client_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansoncht/flight-microservices/internal/reader/client"
	"github.com/ansoncht/flight-microservices/internal/reader/config"
	"github.com/ansoncht/flight-microservices/internal/reader/model"
	"github.com/stretchr/testify/require"
)

func TestNewFlightAPI_ValidConfig_ShouldSucceed(t *testing.T) {
	cfg := config.FlightAPIConfig{
		URL:  "http://localhost:8080",
		User: "testuser",
		Pass: "testpass",
	}

	actual, err := client.NewFlightAPI(cfg, &http.Client{})

	require.NoError(t, err)
	require.NotNil(t, actual)
}

func TestNewFlightAPI_InvalidConfig_ShouldError(t *testing.T) {
	httpClient := &http.Client{}
	tests := []struct {
		cfg     config.FlightAPIConfig
		client  *http.Client
		wantErr string
	}{
		{
			cfg:     config.FlightAPIConfig{URL: "http://x", User: "u", Pass: "p"},
			client:  nil,
			wantErr: "http client is nil",
		},
		{
			cfg:     config.FlightAPIConfig{URL: "", User: "u", Pass: "p"},
			client:  httpClient,
			wantErr: "flight api url is empty",
		},
		{
			cfg:     config.FlightAPIConfig{URL: "http://x", User: "", Pass: "p"},
			client:  httpClient,
			wantErr: "flight api user is empty",
		},
		{
			cfg:     config.FlightAPIConfig{URL: "http://x", User: "u", Pass: ""},
			client:  httpClient,
			wantErr: "flight api password is empty",
		},
	}

	for _, tt := range tests {
		actual, err := client.NewFlightAPI(tt.cfg, tt.client)

		require.Nil(t, actual)
		require.ErrorContains(t, err, tt.wantErr)
	}
}

func TestFetchFlights_ValidArgs_ShouldSucceed(t *testing.T) {
	expected := []model.Flight{{Origin: "VHHH", Destination: "RJTT", Callsign: "CRK452"}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	cfg := config.FlightAPIConfig{
		URL:  server.URL,
		User: "testuser",
		Pass: "testpass",
	}

	client, err := client.NewFlightAPI(cfg, server.Client())

	require.NoError(t, err)
	require.NotNil(t, client)

	actual, err := client.FetchFlights(context.Background(), "VHHH", "1", "2")

	require.NoError(t, err)
	require.NotNil(t, actual)
	require.Len(t, actual, 1)
	require.Equal(t, expected, actual)
}

func TestFetchFlights_InvalidArgs_ShouldError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	cfg := config.FlightAPIConfig{
		URL:  server.URL,
		User: "testuser",
		Pass: "testpass",
	}

	client, err := client.NewFlightAPI(cfg, server.Client())

	require.NoError(t, err)
	require.NotNil(t, client)

	actual, err := client.FetchFlights(context.Background(), "VHHH", "-a", "-b")

	require.ErrorContains(t, err, "unexpected status code")
	require.Nil(t, actual)
}

func TestFetchFlights_HTTPClientError_ShouldError(t *testing.T) {
	t.Run("HTTP Client Error", func(t *testing.T) {
		cfg := config.FlightAPIConfig{
			URL:  "not a valid url",
			User: "user",
			Pass: "pass",
		}

		client, err := client.NewFlightAPI(cfg, &http.Client{})

		require.NoError(t, err)
		require.NotNil(t, client)

		actual, err := client.FetchFlights(context.Background(), "VHHH", "1", "2")

		require.ErrorContains(t, err, "failed to fetch flight")
		require.Nil(t, actual)
	})
}

func TestFetchFlights_InvalidJSON_ShouldError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("{invalid json")); err != nil {
			slog.Error("Failed to write HTTP response", "error", err)
		}
	}))
	defer server.Close()

	cfg := config.FlightAPIConfig{
		URL:  server.URL,
		User: "testuser",
		Pass: "testpass",
	}

	client, err := client.NewFlightAPI(cfg, server.Client())

	require.NoError(t, err)
	require.NotNil(t, client)

	actual, err := client.FetchFlights(context.Background(), "VHHH", "1", "2")

	require.ErrorContains(t, err, "failed to read response body")
	require.Nil(t, actual)
}

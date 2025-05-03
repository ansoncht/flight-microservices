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

func TestNewFlightAPIClient_ValidConfig_ShouldSucceed(t *testing.T) {
	t.Run("Valid Config", func(t *testing.T) {
		cfg := config.FlightAPIClientConfig{
			URL:  "http://localhost:8080",
			User: "testuser",
			Pass: "testpass",
		}

		actual, err := client.NewFlightAPIClient(cfg, &http.Client{})

		require.NoError(t, err)
		require.NotNil(t, actual)
	})
}

func TestNewFlightAPIClient_InvalidConfig_ShouldError(t *testing.T) {
	httpClient := &http.Client{}
	tests := []struct {
		name    string
		cfg     config.FlightAPIClientConfig
		client  *http.Client
		wantErr string
	}{
		{
			name:    "Nil HTTP Client",
			cfg:     config.FlightAPIClientConfig{URL: "http://x", User: "u", Pass: "p"},
			client:  nil,
			wantErr: "http client is nil",
		},
		{
			name:    "Empty URL",
			cfg:     config.FlightAPIClientConfig{URL: "", User: "u", Pass: "p"},
			client:  httpClient,
			wantErr: "flight api url is empty",
		},
		{
			name:    "Empty User",
			cfg:     config.FlightAPIClientConfig{URL: "http://x", User: "", Pass: "p"},
			client:  httpClient,
			wantErr: "flight api user is empty",
		},
		{
			name:    "Empty Password",
			cfg:     config.FlightAPIClientConfig{URL: "http://x", User: "u", Pass: ""},
			client:  httpClient,
			wantErr: "flight api password is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := client.NewFlightAPIClient(tt.cfg, tt.client)

			require.Nil(t, actual)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestFetchFlights_ValidArgs_ShouldSucceed(t *testing.T) {
	t.Run("Valid Args", func(t *testing.T) {
		expected := []model.Flight{{Origin: "VHHH", Destination: "RJTT", Callsign: "CRK452"}}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(expected)
		}))
		defer server.Close()

		cfg := config.FlightAPIClientConfig{
			URL:  server.URL,
			User: "testuser",
			Pass: "testpass",
		}

		client, err := client.NewFlightAPIClient(cfg, server.Client())

		require.NoError(t, err)
		require.NotNil(t, client)

		actual, err := client.FetchFlights(context.Background(), "VHHH", "1", "2")

		require.NoError(t, err)
		require.NotNil(t, actual)
		require.Len(t, actual, 1)
		require.Equal(t, expected, actual)
	})
}

func TestFetchFlights_InvalidArgs_ShouldError(t *testing.T) {
	t.Run("Invalid Args", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "forbidden", http.StatusForbidden)
		}))
		defer server.Close()

		cfg := config.FlightAPIClientConfig{
			URL:  server.URL,
			User: "testuser",
			Pass: "testpass",
		}

		client, err := client.NewFlightAPIClient(cfg, server.Client())

		require.NoError(t, err)
		require.NotNil(t, client)

		actual, err := client.FetchFlights(context.Background(), "VHHH", "-a", "-b")

		require.ErrorContains(t, err, "unexpected status code")
		require.Nil(t, actual)
	})
}

func TestFetchFlights_HTTPClientError_ShouldError(t *testing.T) {
	t.Run("HTTP Client Error", func(t *testing.T) {
		cfg := config.FlightAPIClientConfig{
			URL:  "not a valid url",
			User: "user",
			Pass: "pass",
		}

		client, err := client.NewFlightAPIClient(cfg, &http.Client{})

		require.NoError(t, err)
		require.NotNil(t, client)

		actual, err := client.FetchFlights(context.Background(), "VHHH", "1", "2")

		require.ErrorContains(t, err, "failed to fetch flight")
		require.Nil(t, actual)
	})
}

func TestFetchFlights_InvalidJSON_ShouldError(t *testing.T) {
	t.Run("Invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if _, err := w.Write([]byte("{invalid json")); err != nil {
				slog.Error("Failed to write HTTP response", "error", err)
			}
		}))
		defer server.Close()

		cfg := config.FlightAPIClientConfig{
			URL:  server.URL,
			User: "testuser",
			Pass: "testpass",
		}

		client, err := client.NewFlightAPIClient(cfg, server.Client())

		require.NoError(t, err)
		require.NotNil(t, client)

		actual, err := client.FetchFlights(context.Background(), "VHHH", "1", "2")

		require.ErrorContains(t, err, "failed to read response body")
		require.Nil(t, actual)
	})
}

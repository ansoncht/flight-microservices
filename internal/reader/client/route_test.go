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

func TestNewRouteAPIClient_ValidConfig_ShouldSucceed(t *testing.T) {
	t.Run("Valid Config", func(t *testing.T) {
		cfg := config.RouteAPIClientConfig{
			URL: "http://localhost:8080",
		}

		actual, err := client.NewRouteAPIClient(cfg, &http.Client{})

		require.NoError(t, err)
		require.NotNil(t, actual)
	})
}

func TestNewRouteAPIClient_InvalidConfig_ShouldError(t *testing.T) {
	httpClient := &http.Client{}
	tests := []struct {
		name    string
		cfg     config.RouteAPIClientConfig
		client  *http.Client
		wantErr string
	}{
		{
			name:    "Nil HTTP Client",
			cfg:     config.RouteAPIClientConfig{URL: "http://x"},
			client:  nil,
			wantErr: "http client is nil",
		},
		{
			name:    "Empty URL",
			cfg:     config.RouteAPIClientConfig{URL: ""},
			client:  httpClient,
			wantErr: "route api url is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := client.NewRouteAPIClient(tt.cfg, tt.client)

			require.Nil(t, actual)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestFetchRoute_ValidArgs_ShouldSucceed(t *testing.T) {
	t.Run("Valid Args", func(t *testing.T) {
		expected := model.Route{
			Response: model.Response{
				FlightRoute: model.FlightRoute{
					CallSign: "CRK452",
					Destination: model.Airport{
						ICAOCode: "VHHH",
					},
					Origin: model.Airport{
						ICAOCode: "RJTT",
					},
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(expected)
		}))
		defer server.Close()

		cfg := config.RouteAPIClientConfig{
			URL: server.URL,
		}

		client, err := client.NewRouteAPIClient(cfg, server.Client())

		require.NoError(t, err)
		require.NotNil(t, client)

		actual, err := client.FetchRoute(context.Background(), "CRK452")

		require.NoError(t, err)
		require.NotNil(t, actual)
		require.Equal(t, expected, *actual)
	})
}

func TestFetchRoute_InvalidArgs_ShouldError(t *testing.T) {
	t.Run("Invalid Args", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "forbidden", http.StatusForbidden)
		}))
		defer server.Close()

		cfg := config.RouteAPIClientConfig{
			URL: server.URL,
		}

		client, err := client.NewRouteAPIClient(cfg, server.Client())

		require.NoError(t, err)
		require.NotNil(t, client)

		actual, err := client.FetchRoute(context.Background(), "ABC123")

		require.ErrorContains(t, err, "unexpected status code")
		require.Nil(t, actual)
	})
}

func TestFetchRoute_HTTPClientError_ShouldError(t *testing.T) {
	t.Run("HTTP Client Error", func(t *testing.T) {
		cfg := config.RouteAPIClientConfig{
			URL: "not a valid url",
		}

		client, err := client.NewRouteAPIClient(cfg, &http.Client{})

		require.NoError(t, err)
		require.NotNil(t, client)

		actual, err := client.FetchRoute(context.Background(), "CRK452")

		require.ErrorContains(t, err, "failed to fetch route")
		require.Nil(t, actual)
	})
}

func TestFetchRoute_InvalidJSON_ShouldError(t *testing.T) {
	t.Run("Invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if _, err := w.Write([]byte("{invalid json")); err != nil {
				slog.Error("Failed to write HTTP response", "error", err)
			}
		}))
		defer server.Close()

		cfg := config.RouteAPIClientConfig{
			URL: server.URL,
		}

		client, err := client.NewRouteAPIClient(cfg, server.Client())

		require.NoError(t, err)
		require.NotNil(t, client)

		actual, err := client.FetchRoute(context.Background(), "CRK452")

		require.ErrorContains(t, err, "failed to read response body")
		require.Nil(t, actual)
	})
}

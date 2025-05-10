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

func TestNewRouteAPI_ValidConfig_ShouldSucceed(t *testing.T) {
	cfg := config.RouteAPIConfig{
		URL: "http://localhost:8080",
	}
	client, err := client.NewRouteAPI(cfg, &http.Client{})
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewRouteAPI_InvalidConfig_ShouldError(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.RouteAPIConfig
		client  *http.Client
		wantErr string
	}{
		{
			name:    "Nil HTTP Client",
			cfg:     config.RouteAPIConfig{URL: "http://x"},
			client:  nil,
			wantErr: "http client is nil",
		},
		{
			name:    "Empty URL",
			cfg:     config.RouteAPIConfig{URL: ""},
			client:  &http.Client{},
			wantErr: "route api url is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := client.NewRouteAPI(tt.cfg, tt.client)
			require.Nil(t, client)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestFetchRoute_ValidArgs_ShouldSucceed(t *testing.T) {
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

	cfg := config.RouteAPIConfig{
		URL: server.URL,
	}
	client, err := client.NewRouteAPI(cfg, server.Client())
	require.NoError(t, err)
	require.NotNil(t, client)

	route, err := client.FetchRoute(context.Background(), "CRK452")
	require.NoError(t, err)
	require.NotNil(t, route)
	require.Equal(t, expected, *route)
}

func TestFetchRoute_InvalidArgs_ShouldError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	cfg := config.RouteAPIConfig{
		URL: server.URL,
	}
	client, err := client.NewRouteAPI(cfg, server.Client())
	require.NoError(t, err)
	require.NotNil(t, client)

	route, err := client.FetchRoute(context.Background(), "ABC123")
	require.ErrorContains(t, err, "unexpected status code")
	require.Nil(t, route)
}

func TestFetchRoute_HTTPClientError_ShouldError(t *testing.T) {
	cfg := config.RouteAPIConfig{
		URL: "not a valid url",
	}

	client, err := client.NewRouteAPI(cfg, &http.Client{})
	require.NoError(t, err)
	require.NotNil(t, client)

	route, err := client.FetchRoute(context.Background(), "CRK452")
	require.ErrorContains(t, err, "failed to fetch route")
	require.Nil(t, route)
}

func TestFetchRoute_InvalidJSON_ShouldError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("{invalid json")); err != nil {
			slog.Error("Failed to write HTTP response", "error", err)
		}
	}))
	defer server.Close()

	cfg := config.RouteAPIConfig{
		URL: server.URL,
	}
	client, err := client.NewRouteAPI(cfg, server.Client())
	require.NoError(t, err)
	require.NotNil(t, client)

	route, err := client.FetchRoute(context.Background(), "CRK452")
	require.ErrorContains(t, err, "failed to read response body")
	require.Nil(t, route)
}

func TestFetchRoute_InvalidURL_ShouldError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("{invalid json")); err != nil {
			slog.Error("Failed to write HTTP response", "error", err)
		}
	}))
	defer server.Close()

	cfg := config.RouteAPIConfig{
		URL: "http://example.com:abc",
	}
	client, err := client.NewRouteAPI(cfg, server.Client())
	require.NoError(t, err)
	require.NotNil(t, client)

	route, err := client.FetchRoute(context.Background(), "CRK452")
	require.ErrorContains(t, err, "failed to parse url")
	require.Nil(t, route)
}

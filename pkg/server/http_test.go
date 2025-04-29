package server_test

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ansoncht/flight-microservices/pkg/server"
	"github.com/stretchr/testify/require"
)

func testHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNewHTTPServer_ValidConfigAndHandler_ShouldSucceed(t *testing.T) {
	t.Run("Valid Config", func(t *testing.T) {
		cfg := server.HTTPConfig{Port: "8080", Timeout: 5}
		actual, err := server.NewHTTPServer(cfg, http.HandlerFunc(testHandler))

		require.NoError(t, err)
		require.NotNil(t, actual)
	})
}

func TestNewHTTPServer_InvalidPort_ShouldError(t *testing.T) {
	tests := []struct {
		name    string
		cfg     server.HTTPConfig
		wantErr string
	}{
		{
			name:    "Empty Port",
			cfg:     server.HTTPConfig{Port: "", Timeout: 5},
			wantErr: "port number is empty",
		},
		{
			name:    "Invalid Port",
			cfg:     server.HTTPConfig{Port: "abc", Timeout: 5},
			wantErr: "port number is invalid",
		},
		{
			name:    "Negative Port",
			cfg:     server.HTTPConfig{Port: "-1010", Timeout: 5},
			wantErr: "port number must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := server.NewHTTPServer(tt.cfg, http.HandlerFunc(testHandler))

			require.Nil(t, actual)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestNewHTTPServer_InvalidTimeout_ShouldError(t *testing.T) {
	tests := []struct {
		name string
		cfg  server.HTTPConfig
	}{
		{
			name: "Zero Timeout",
			cfg:  server.HTTPConfig{Port: "8080", Timeout: 0},
		},
		{
			name: "Negative Timeout",
			cfg:  server.HTTPConfig{Port: "8080", Timeout: -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := server.NewHTTPServer(tt.cfg, http.HandlerFunc(testHandler))

			require.Nil(t, actual)
			require.ErrorContains(t, err, "http server timeout is invalid")
		})
	}
}

func TestNewHTTPServer_NilHandler_ShouldError(t *testing.T) {
	t.Run("Nil Handler", func(t *testing.T) {
		cfg := server.HTTPConfig{Port: "8080", Timeout: 5}
		actual, err := server.NewHTTPServer(cfg, nil)

		require.Nil(t, actual)
		require.ErrorContains(t, err, "handler is nil")
	})
}

func TestServe_ContextCanceledOrDeadlineExceeded_ShouldError(t *testing.T) {
	cfg := server.HTTPConfig{Port: "8081", Timeout: 2}
	actual, err := server.NewHTTPServer(cfg, http.HandlerFunc(testHandler))

	require.NoError(t, err)
	require.NotNil(t, actual)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	errCh := make(chan error)
	go func() {
		errCh <- actual.Serve(ctx)
	}()

	err = <-errCh

	require.ErrorContains(t, err, "context canceled while running HTTP server")
}

func TestServe_ServerError_ShouldError(t *testing.T) {
	ln, err := net.Listen("tcp", "localhost:0")

	require.NoError(t, err)

	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port

	cfg := server.HTTPConfig{Port: fmt.Sprintf("%d", port), Timeout: 5}
	actual, err := server.NewHTTPServer(cfg, http.HandlerFunc(testHandler))

	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = actual.Serve(ctx)

	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "context canceled while running HTTP server") ||
			strings.Contains(err.Error(), "failed to start HTTP server"),
	)
}

func TestClose_GracefulShutdown_ShouldSucceed(t *testing.T) {
	cfg := server.HTTPConfig{Port: "8082", Timeout: 5}
	actual, err := server.NewHTTPServer(cfg, http.HandlerFunc(testHandler))

	require.NoError(t, err)
	require.NotNil(t, actual)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = actual.Serve(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Close the server
	err = actual.Close(ctx)

	require.NoError(t, err)
}

func TestClose_ContextCanceledOrDeadlineExceeded_ShouldError(t *testing.T) {
	cfg := server.HTTPConfig{Port: "8086", Timeout: 5}

	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	actual, err := server.NewHTTPServer(cfg, slowHandler)

	require.NoError(t, err)
	require.NotNil(t, actual)

	go func() {
		_ = actual.Serve(context.Background())
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	go func() {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:8086", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Error("Failed to perform HTTP request", "error", err)
		}

		defer resp.Body.Close()
	}()

	time.Sleep(200 * time.Millisecond)

	// Use a context that will expire very soon
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	err = actual.Close(ctx)

	require.ErrorContains(t, err, "failed to shutdown")
}

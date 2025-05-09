package http_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	server "github.com/ansoncht/flight-microservices/pkg/http"
	"github.com/stretchr/testify/require"
)

func testHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestNewHTTPServer_ValidConfigAndHandler_ShouldSucceed(t *testing.T) {
	cfg := server.ServerConfig{Port: "8080", Timeout: 5}
	server, err := server.NewServer(cfg, http.HandlerFunc(testHandler))
	require.NoError(t, err)
	require.NotNil(t, server)
}

func TestNewHTTPServer_InvalidConfig_ShouldError(t *testing.T) {
	tests := []struct {
		name    string
		cfg     server.ServerConfig
		handler http.Handler
		wantErr string
	}{
		{
			name:    "Empty Port",
			cfg:     server.ServerConfig{Port: "", Timeout: 5},
			handler: http.HandlerFunc(testHandler),
			wantErr: "port number is empty",
		},
		{
			name:    "Invalid Port",
			cfg:     server.ServerConfig{Port: "abc", Timeout: 5},
			handler: http.HandlerFunc(testHandler),

			wantErr: "port number is invalid",
		},
		{
			name:    "Negative Port",
			cfg:     server.ServerConfig{Port: "-1010", Timeout: 5},
			handler: http.HandlerFunc(testHandler),

			wantErr: "port number must be greater than 0",
		},
		{
			name:    "Zero Timeout",
			cfg:     server.ServerConfig{Port: "8080", Timeout: 0},
			handler: http.HandlerFunc(testHandler),

			wantErr: "http server timeout is invalid",
		},
		{
			name:    "Negative Timeout",
			cfg:     server.ServerConfig{Port: "8080", Timeout: -1},
			handler: http.HandlerFunc(testHandler),

			wantErr: "http server timeout is invalid",
		},
		{
			name:    "Nil Handler",
			cfg:     server.ServerConfig{Port: "8080", Timeout: 5},
			handler: nil,
			wantErr: "handler is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := server.NewServer(tt.cfg, tt.handler)
			require.Nil(t, server)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestServe_ContextCanceledOrDeadlineExceeded_ShouldError(t *testing.T) {
	cfg := server.ServerConfig{Port: "8081", Timeout: 2}
	server, err := server.NewServer(cfg, http.HandlerFunc(testHandler))
	require.NoError(t, err)
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	errCh := make(chan error)
	go func() {
		errCh <- server.Serve(ctx)
	}()

	err = <-errCh
	require.ErrorContains(t, err, "context canceled while running HTTP server")
}

func TestServe_ServerError_ShouldError(t *testing.T) {
	ln, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer func() {
		err := ln.Close()
		require.NoError(t, err)
	}()

	cfg := server.ServerConfig{Port: fmt.Sprintf("%d", 123), Timeout: 5}
	server, err := server.NewServer(cfg, http.HandlerFunc(testHandler))
	require.NoError(t, err)

	err = server.Serve(context.Background())

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to start HTTP server")
}

func TestClose_GracefulShutdown_ShouldSucceed(t *testing.T) {
	cfg := server.ServerConfig{Port: "8082", Timeout: 5}
	server, err := server.NewServer(cfg, http.HandlerFunc(testHandler))
	require.NoError(t, err)
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = server.Serve(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Close the server
	err = server.Close(ctx)
	require.NoError(t, err)
}

func TestClose_ContextCanceledOrDeadlineExceeded_ShouldError(t *testing.T) {
	cfg := server.ServerConfig{Port: "8086", Timeout: 5}
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	})
	server, err := server.NewServer(cfg, slowHandler)
	require.NoError(t, err)
	require.NotNil(t, server)

	go func() {
		_ = server.Serve(context.Background())
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	errChan := make(chan error)
	go func() {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:8086", nil)
		errChan <- err
		resp, err := http.DefaultClient.Do(req)
		errChan <- err
		defer resp.Body.Close()
	}()
	require.NoError(t, <-errChan)

	time.Sleep(200 * time.Millisecond)

	// Use a context that will expire very soon
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	err = server.Close(ctx)
	require.ErrorContains(t, err, "failed to shutdown")
}

package http_test

import (
	"testing"
	"time"

	"github.com/ansoncht/flight-microservices/pkg/http"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient_ValidTimeout_ShouldSucceed(t *testing.T) {
	cfg := http.ClientConfig{Timeout: 5}
	actual, err := http.NewClient(cfg)

	require.NoError(t, err)
	require.NotNil(t, actual)
	require.Equal(t, time.Duration(cfg.Timeout)*time.Second, actual.Timeout)
}

func TestNewHTTPClient_InvalidTimeout_ShouldError(t *testing.T) {
	tests := []struct {
		cfg http.ClientConfig
	}{
		{
			cfg: http.ClientConfig{Timeout: 0},
		},
		{
			cfg: http.ClientConfig{Timeout: -1},
		},
	}

	for _, tt := range tests {
		actual, err := http.NewClient(tt.cfg)

		require.Nil(t, actual)
		require.ErrorContains(t, err, "http client timeout is invalid")
	}
}

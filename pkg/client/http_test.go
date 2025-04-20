package client_test

import (
	"testing"
	"time"

	"github.com/ansoncht/flight-microservices/pkg/client"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient_ValidTimeout_ShouldSucceed(t *testing.T) {
	t.Run("Valid Timeout", func(t *testing.T) {
		cfg := client.HTTPConfig{Timeout: 5}
		actual, err := client.NewHTTPClient(cfg)

		require.NoError(t, err)
		require.NotNil(t, actual)
		require.Equal(t, time.Duration(cfg.Timeout)*time.Second, actual.Timeout)
	})
}

func TestNewHTTPClient_InvalidTimeout_ShouldError(t *testing.T) {
	tests := []struct {
		name string
		cfg  client.HTTPConfig
	}{
		{
			name: "Zero Timeout",
			cfg:  client.HTTPConfig{Timeout: 0},
		},
		{
			name: "Negative Timeout",
			cfg:  client.HTTPConfig{Timeout: -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := client.NewHTTPClient(tt.cfg)

			require.Nil(t, actual)
			require.ErrorContains(t, err, "http client timeout is invalid")
		})
	}
}

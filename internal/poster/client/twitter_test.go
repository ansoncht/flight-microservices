package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/ansoncht/flight-microservices/internal/poster/client"
	"github.com/ansoncht/flight-microservices/internal/poster/config"
	"github.com/stretchr/testify/require"
)

func TestNewTwitterAPI_ValidConfig_ShouldSucceed(t *testing.T) {
	t.Setenv("GOTWI_API_KEY", "test")
	t.Setenv("GOTWI_API_KEY_SECRET", "test")

	cfg := config.TwitterAPIConfig{
		Key:    os.Getenv("GOTWI_API_KEY"),
		Secret: os.Getenv("GOTWI_API_KEY_SECRET"),
	}
	client, err := client.NewTwitterAPI(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewTwitterAPI_InvalidConfig_ShouldError(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.TwitterAPIConfig
		wantErr string
	}{
		{
			name:    "Empty Key",
			cfg:     config.TwitterAPIConfig{Key: "", Secret: "test"},
			wantErr: "twitter api key is empty",
		},
		{
			name:    "Empty Secret",
			cfg:     config.TwitterAPIConfig{Key: "test", Secret: ""},
			wantErr: "twitter api secret is empty",
		},
		{
			name:    "Empty Environment Variables",
			cfg:     config.TwitterAPIConfig{Key: "test", Secret: "test"},
			wantErr: "failed to create twiiter api wrapper client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := client.NewTwitterAPI(tt.cfg)
			require.Nil(t, client)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestPublishPost_EmptyContent_ShouldError(t *testing.T) {
	t.Setenv("GOTWI_API_KEY", "test")
	t.Setenv("GOTWI_API_KEY_SECRET", "test")

	cfg := config.TwitterAPIConfig{
		Key:    os.Getenv("GOTWI_API_KEY"),
		Secret: os.Getenv("GOTWI_API_KEY_SECRET"),
	}
	client, err := client.NewTwitterAPI(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx := context.Background()
	err = client.PublishPost(ctx, "")
	require.ErrorContains(t, err, "content is empty")
}

func TestPublishPost_TwitterWrapperError_ShouldError(t *testing.T) {
	t.Setenv("GOTWI_API_KEY", "test")
	t.Setenv("GOTWI_API_KEY_SECRET", "test")

	cfg := config.TwitterAPIConfig{
		Key:    os.Getenv("GOTWI_API_KEY"),
		Secret: os.Getenv("GOTWI_API_KEY_SECRET"),
	}
	client, err := client.NewTwitterAPI(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx := context.Background()
	err = client.PublishPost(ctx, "test post ")
	require.ErrorContains(t, err, "failed to create Twitter post")
}

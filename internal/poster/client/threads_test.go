package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ansoncht/flight-microservices/internal/poster/client"
	"github.com/ansoncht/flight-microservices/internal/poster/config"
	"github.com/ansoncht/flight-microservices/internal/poster/model"
	"github.com/stretchr/testify/require"
)

func TestNewThreadsAPI_ValidConfig_ShouldSucceed(t *testing.T) {
	expected := model.ThreadsUserResponse{
		ID: "testuser",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(expected)
	}))

	ctx := context.Background()
	cfg := config.ThreadsAPIConfig{
		URL:   server.URL,
		Token: "test",
	}
	client, err := client.NewThreadsAPI(ctx, cfg, &http.Client{})
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewThreadsAPI_InvalidConfig_ShouldError(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.ThreadsAPIConfig
		client  *http.Client
		wantErr string
	}{
		{
			name:    "Nil HTTP Client",
			cfg:     config.ThreadsAPIConfig{URL: "http://localhost:8080", Token: "test"},
			client:  nil,
			wantErr: "http client is nil",
		},
		{
			name:    "Empty URL",
			cfg:     config.ThreadsAPIConfig{URL: "", Token: "test"},
			client:  &http.Client{},
			wantErr: "threads api url is empty",
		},
		{
			name:    "Empty Token",
			cfg:     config.ThreadsAPIConfig{URL: "http://localhost:8080", Token: ""},
			client:  &http.Client{},
			wantErr: "threads api token is empty",
		},
		{
			name:    "Empty User",
			cfg:     config.ThreadsAPIConfig{URL: "http://localhost:8080", Token: "test"},
			client:  &http.Client{},
			wantErr: "failed to get user id",
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := client.NewThreadsAPI(ctx, tt.cfg, tt.client)
			require.Nil(t, client)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestPublishPost_ValidArgs_ShouldSucceed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(&model.ThreadsPostResponse{
			ID: "testpost",
		})
	}))
	defer server.Close()

	ctx := context.Background()
	cfg := config.ThreadsAPIConfig{
		URL:   server.URL,
		Token: "test",
	}
	client, err := client.NewThreadsAPI(ctx, cfg, server.Client())
	require.NoError(t, err)
	require.NotNil(t, client)

	err = client.PublishPost(ctx, "Test post")
	require.NoError(t, err)
}

func TestPublishPost_ContainerFailure_ShouldError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(&model.ThreadsUserResponse{
			ID: "testuser",
		})
	}))

	defer server.Close()

	ctx := context.Background()
	cfg := config.ThreadsAPIConfig{
		URL:   server.URL,
		Token: "test",
	}
	client, err := client.NewThreadsAPI(ctx, cfg, server.Client())
	require.NoError(t, err)
	require.NotNil(t, client)

	err = client.PublishPost(ctx, "")
	require.ErrorContains(t, err, "content is empty")
}

func TestPublishPost_InvalidContainerResponse_ShouldError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/me":
			_ = json.NewEncoder(w).Encode(&model.ThreadsUserResponse{
				ID: "testuser",
			})
		case "/testuser/threads":
			_ = json.NewEncoder(w).Encode([]byte("{invalid json"))
		default:
			http.NotFound(w, r)
		}
	}))

	defer server.Close()

	ctx := context.Background()
	cfg := config.ThreadsAPIConfig{
		URL:   server.URL,
		Token: "test",
	}
	client, err := client.NewThreadsAPI(ctx, cfg, server.Client())
	require.NoError(t, err)
	require.NotNil(t, client)

	err = client.PublishPost(ctx, "Test post")
	require.ErrorContains(t, err, "failed to decode Threads container response")
}

func TestPublishPost_InvalidPostResponse_ShouldError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/me":
			_ = json.NewEncoder(w).Encode(&model.ThreadsUserResponse{
				ID: "testuser",
			})
		case "/testuser/threads":
			_ = json.NewEncoder(w).Encode(&model.ThreadsContainerResponse{
				ID: "testcontainer",
			})
		case "/testuser/threads_publish":
			_ = json.NewEncoder(w).Encode([]byte("{invalid json"))
		}
	}))
	defer server.Close()

	ctx := context.Background()
	cfg := config.ThreadsAPIConfig{
		URL:   server.URL,
		Token: "test",
	}
	client, err := client.NewThreadsAPI(ctx, cfg, server.Client())
	require.NoError(t, err)
	require.NotNil(t, client)

	err = client.PublishPost(ctx, "Test post")
	require.ErrorContains(t, err, "failed to decode Threads post response")
}

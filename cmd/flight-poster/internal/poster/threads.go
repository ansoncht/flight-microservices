package poster

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/model"
)

type ThreadsClient struct {
	Token      string
	User       string
	BaseURL    string
	httpClient *http.Client
}

func NewThreadsClient(ctx context.Context, httpClient *http.Client) (*ThreadsClient, error) {
	slog.Info("Creating Threads client for the service")

	cfg, err := config.LoadThreadsClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load mongo client config: %w", err)
	}

	client := &ThreadsClient{
		BaseURL:    cfg.URL,
		Token:      cfg.Token,
		httpClient: httpClient,
	}

	err = client.refreshToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get / refresh Threads user token: %w", err)
	}

	user, err := client.getUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Threads user id: %w", err)
	}

	client.User = user

	return client, nil
}

func (t *ThreadsClient) CreatePost(content string, media []string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("failed to create thread: content is empty")
	}

	if len(media) == 0 {
		slog.Warn("Thread contains no media")
	}

	return content, nil
}
func (t *ThreadsClient) PublishPost(postID string) (bool, error) {
	if postID == "" {
		return false, fmt.Errorf("failed to post thread: post id is empty")
	}
	return true, nil
}

func (t *ThreadsClient) getUserID(ctx context.Context) (string, error) {
	// Parse the base URL
	endpoint, err := url.Parse(t.BaseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}

	// Add path segments
	endpoint = endpoint.JoinPath("me")

	// Add and set query parameters
	query := endpoint.Query()
	query.Add("access_token", t.Token)

	// Create a HTTP GET request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for flight: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get user id: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var user model.ThreadsUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to parse user: %w", err)
	}

	return user.ID, nil
}

func (t *ThreadsClient) refreshToken(ctx context.Context) error {
	// Parse the base URL
	endpoint, err := url.Parse(t.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	// Add path segments
	endpoint = endpoint.JoinPath("refresh_access_token")

	// Add and set query parameters
	query := endpoint.Query()
	query.Add("grant_type", "th_refresh_token")
	query.Add("access_token", t.Token)
	endpoint.RawQuery = query.Encode()

	// Create a HTTP GET request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request for refreshing token: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

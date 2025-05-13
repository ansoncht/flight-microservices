package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/ansoncht/flight-microservices/internal/poster/config"
	"github.com/ansoncht/flight-microservices/internal/poster/model"
)

type Threads struct {
	token   token
	user    string
	baseURL string
	client  *http.Client
}

func NewThreadsAPI(
	ctx context.Context,
	cfg config.ThreadsAPIConfig,
	client *http.Client,
) (*Threads, error) {
	slog.Info("Initializing Threads API client", "url", cfg.URL)

	if client == nil {
		return nil, fmt.Errorf("http client is nil")
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("threads api url is empty")
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("threads api token is empty")
	}

	// Assumes initial token expires after 60 days
	token := token{
		accessToken: cfg.Token,
		expiration:  time.Now().Add(60 * 24 * time.Hour),
	}

	threads := &Threads{
		baseURL: cfg.URL,
		token:   token,
		client:  client,
	}

	user, err := threads.getUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user id: %w", err)
	}

	threads.user = user

	return threads, nil
}

func (t *Threads) PublishPost(ctx context.Context, content string) error {
	postID, err := t.createContainer(ctx, content, nil)
	if err != nil {
		return fmt.Errorf("failed to create Threads container: %w", err)
	}

	if t.needRefreshToken() {
		if err := t.refreshToken(ctx); err != nil {
			return fmt.Errorf("failed to refresh Threads user token: %w", err)
		}
	}

	// Parse the base URL
	endpoint, err := url.Parse(t.baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	// Add path segments
	endpoint = endpoint.JoinPath(t.user, "threads_publish")

	// Add and set query parameters
	query := endpoint.Query()
	query.Add("creation_id", postID)
	query.Add("access_token", t.token.accessToken)
	endpoint.RawQuery = query.Encode()

	// Create a HTTP POST request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var post model.ThreadsPostResponse
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return fmt.Errorf("failed to decode Threads post response: %w", err)
	}

	return nil
}

func (t *Threads) createContainer(ctx context.Context, content string, media []string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("content is empty")
	}

	if len(media) != 0 {
		return "", fmt.Errorf("media is not supported")
	}

	// Parse the base URL
	endpoint, err := url.Parse(t.baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}

	// Add path segments
	endpoint = endpoint.JoinPath(t.user, "threads")

	// Add and set query parameters
	query := endpoint.Query()
	query.Add("text", content)
	query.Add("media_type", "TEXT")
	query.Add("access_token", t.token.accessToken)
	endpoint.RawQuery = query.Encode()

	// Create a HTTP POST request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var post model.ThreadsContainerResponse
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return "", fmt.Errorf("failed to decode Threads container response: %w", err)
	}

	return post.ID, nil
}

func (t *Threads) getUserID(ctx context.Context) (string, error) {
	// Parse the base URL
	endpoint, err := url.Parse(t.baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}

	// Add path segments
	endpoint = endpoint.JoinPath("me")

	// Add and set query parameters
	query := endpoint.Query()
	query.Add("access_token", t.token.accessToken)
	endpoint.RawQuery = query.Encode()

	// Create a HTTP GET request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var user model.ThreadsUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to decode Threads user response: %w", err)
	}

	return user.ID, nil
}

func (t *Threads) refreshToken(ctx context.Context) error {
	// Parse the base URL
	endpoint, err := url.Parse(t.baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	// Add path segments
	endpoint = endpoint.JoinPath("refresh_access_token")

	// Add and set query parameters
	query := endpoint.Query()
	query.Add("grant_type", "th_refresh_token")
	query.Add("access_token", t.token.accessToken)
	endpoint.RawQuery = query.Encode()

	// Create a HTTP GET request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	t.token.expiration = time.Now().Add(60 * 24 * time.Hour)

	return nil
}

func (t *Threads) needRefreshToken() bool {
	// If the token's expiration time is within 7 days from now, it's time to refresh
	return time.Now().After(t.token.expiration.Add(-7 * 24 * time.Hour))
}

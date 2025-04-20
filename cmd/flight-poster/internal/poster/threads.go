package poster

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/model"
)

type ThreadsClient struct {
	token      token
	user       string
	baseURL    string
	httpClient *http.Client
}

func NewThreadsClient(
	ctx context.Context,
	cfg config.ThreadsClientConfig,
	httpClient *http.Client,
) (*ThreadsClient, error) {
	slog.Info("Creating Threads client for the service")

	// Assumes initial token expires after 60 days
	token := token{
		accessToken: cfg.Token,
		expiration:  time.Now().Add(60 * 24 * time.Hour),
	}

	client := &ThreadsClient{
		baseURL:    cfg.URL,
		token:      token,
		httpClient: httpClient,
	}

	user, err := client.getUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Threads user id: %w", err)
	}

	client.user = user

	return client, nil
}

func (t *ThreadsClient) PublishPost(ctx context.Context, content string) (bool, error) {
	postID, err := t.createPost(ctx, content, nil)
	if err != nil {
		return false, fmt.Errorf("failed to post Threads post: %w", err)
	}

	if t.needRefreshToken() {
		if err := t.refreshToken(ctx); err != nil {
			return false, fmt.Errorf("failed to refresh Threads user token: %w", err)
		}
	}

	// Parse the base URL
	endpoint, err := url.Parse(t.baseURL)
	if err != nil {
		return false, fmt.Errorf("failed to parse url: %w", err)
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
		return false, fmt.Errorf("failed to create request for posting Threads post : %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to post Threads post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var post model.ThreadsPostCreationResponse
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return false, fmt.Errorf("failed to parse container: %w", err)
	}

	return true, nil
}

func (t *ThreadsClient) createPost(ctx context.Context, content string, media []string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("failed to create Threads post: content is empty")
	}

	if len(media) == 0 {
		slog.Warn("Threads post contains no media")
	}

	if t.needRefreshToken() {
		if err := t.refreshToken(ctx); err != nil {
			return "", fmt.Errorf("failed to refresh Threads user token: %w", err)
		}
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
		return "", fmt.Errorf("failed to create request for initializing Threads post: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create Threads post container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var post model.ThreadsPostCreationResponse
	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return "", fmt.Errorf("failed to parse container: %w", err)
	}

	return post.ID, nil
}

func (t *ThreadsClient) getUserID(ctx context.Context) (string, error) {
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

	t.token.expiration = time.Now().Add(60 * 24 * time.Hour)

	return nil
}

func (t *ThreadsClient) needRefreshToken() bool {
	// If the token's expiration time is within 7 days from now, it's time to refresh
	return time.Now().After(t.token.expiration.Add(-7 * 24 * time.Hour))
}

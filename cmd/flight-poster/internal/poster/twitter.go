package poster

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/config"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/tweet/managetweet/types"
)

type TwitterClient struct {
	client *gotwi.Client
}

func NewTwitterClient() (*TwitterClient, error) {
	slog.Info("Creating Twitter client for the service")

	cfg, err := config.LoadTwitterClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load Twitter client config: %w", err)
	}

	opts := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           cfg.Key,
		OAuthTokenSecret:     cfg.Secret,
	}

	client, err := gotwi.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize api wrapper client: %w", err)
	}

	return &TwitterClient{
		client: client,
	}, nil
}

func (t *TwitterClient) PublishPost(ctx context.Context, content string) (bool, error) {
	in := &types.CreateInput{
		Text: gotwi.String(content),
	}

	res, err := managetweet.Create(ctx, t.client, in)
	if err != nil || res.Data.ID == nil || *res.Data.ID == "" {
		return false, fmt.Errorf("failed to post Twitter post: %w", err)
	}

	return true, nil
}

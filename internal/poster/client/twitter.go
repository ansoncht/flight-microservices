package client

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ansoncht/flight-microservices/internal/poster/config"

	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/tweet/managetweet/types"
)

type Twitter struct {
	client *gotwi.Client
}

func NewTwitterAPI(cfg config.TwitterAPIConfig) (*Twitter, error) {
	slog.Info("Initializing Twitter API client")

	if cfg.Key == "" {
		return nil, fmt.Errorf("twitter api key is empty")
	}

	if cfg.Secret == "" {
		return nil, fmt.Errorf("twitter api secret is empty")
	}

	opts := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           cfg.Key,
		OAuthTokenSecret:     cfg.Secret,
	}

	client, err := gotwi.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create twiiter api wrapper client: %w", err)
	}

	return &Twitter{
		client: client,
	}, nil
}

func (t *Twitter) PublishPost(ctx context.Context, content string) error {
	if content == "" {
		return fmt.Errorf("content is empty")
	}

	in := &types.CreateInput{
		Text: gotwi.String(content),
	}

	res, err := managetweet.Create(ctx, t.client, in)
	if err != nil || res.Data.ID == nil || *res.Data.ID == "" {
		return fmt.Errorf("failed to create Twitter post: %w", err)
	}

	return nil
}

package client

import (
	"context"
	"time"
)

// Socials defines the interface for posting content to social media platforms.
type Socials interface {
	// PublishPost publishes a post to the social media platform.
	PublishPost(ctx context.Context, content string) error
}

// token holds the access token and its expiration time.
type token struct {
	accessToken string    // Actual access token for the user
	expiration  time.Time // Expiration time of the token
}

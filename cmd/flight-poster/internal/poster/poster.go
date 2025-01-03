package poster

import (
	"context"
	"time"
)

type Poster interface {
	CreatePost(ctx context.Context, content string, media []string) (string, error)
	PublishPost(ctx context.Context, postID string) (bool, error)
}

type token struct {
	accessToken string    // Actual access token for the user
	expiration  time.Time // Expiration time of the token
}

package service

import (
	"context"
	"fmt"

	"github.com/ansoncht/flight-microservices/internal/poster/client"
	"github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/ansoncht/flight-microservices/pkg/repository"
	"github.com/twmb/franz-go/pkg/kgo"
	"golang.org/x/sync/errgroup"
)

// Poster holds dependencies for posting flight summaries to social media platforms.
type Poster struct {
	// socials specifies the list of social media clients to post messages.
	socials []client.Socials
	// messageReader specifies the message reader to read messages from a message queue.
	messageReader kafka.MessageReader
	// repo  specifies the repo to interact with the db collection.
	repo repository.SummaryRepository
}

// NewPoster creates a new Poster instance based on the provided social media clients, message reader, and repository.
func NewPoster(
	socials []client.Socials,
	messageReader kafka.MessageReader,
	repo repository.SummaryRepository,
) (*Poster, error) {
	if len(socials) == 0 {
		return nil, fmt.Errorf("social media clients are empty")
	}

	if messageReader == nil {
		return nil, fmt.Errorf("message reader is nil")
	}

	if repo == nil {
		return nil, fmt.Errorf("repository is nil")
	}

	return &Poster{
		socials:       socials,
		messageReader: messageReader,
		repo:          repo,
	}, nil
}

// Close closes the poster service.
func (p *Poster) Close() {
	p.messageReader.Close()
}

// Post posts the flight summary to all social media clients.
func (p *Poster) Post(ctx context.Context) error {
	msgChan := make(chan kgo.Record)
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return p.messageReader.ReadMessages(gCtx, msgChan)
	})

postingLoop:
	for {
		select {
		case <-gCtx.Done():
			break postingLoop
		case msg, ok := <-msgChan:
			if !ok {
				break postingLoop
			}

			summary, err := p.repo.Get(gCtx, string(msg.Value))
			if err != nil {
				return fmt.Errorf("failed to get flight summary: %w", err)
			}

			content := summary.FormatForSocialMedia()

			for _, social := range p.socials {
				platform := social
				g.Go(func() error {
					if err := platform.PublishPost(gCtx, content); err != nil {
						return fmt.Errorf("failed to post content: %w", err)
					}

					return nil
				})
			}
		}
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error while reading messages: %w", err)
	}

	if ctx.Err() != nil {
		return fmt.Errorf("context canceled while posting content: %w", ctx.Err())
	}

	return nil
}

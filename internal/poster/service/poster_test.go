package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ansoncht/flight-microservices/internal/poster/client"
	"github.com/ansoncht/flight-microservices/internal/poster/service"
	"github.com/ansoncht/flight-microservices/internal/test/mock"
	"github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/ansoncht/flight-microservices/pkg/model"
	"github.com/ansoncht/flight-microservices/pkg/repository"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/mock/gomock"
)

func TestNewPoster_NonNilClients_ShouldSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	socials := []client.Socials{mock.NewMockSocials(ctrl)}
	reader := mock.NewMockMessageReader(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	poster, err := service.NewPoster(socials, reader, repo)
	require.NoError(t, err)
	require.NotNil(t, poster)
}

func TestNewPoster_NilDependencies_ShouldError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		socials     []client.Socials
		reader      kafka.MessageReader
		repository  repository.SummaryRepository
		expectedErr string
	}{
		{
			name:        "nil socials",
			socials:     nil,
			reader:      mock.NewMockMessageReader(ctrl),
			repository:  mock.NewMockSummaryRepository(ctrl),
			expectedErr: "social media clients are empty",
		},
		{
			name:        "nil reader",
			socials:     []client.Socials{mock.NewMockSocials(ctrl)},
			reader:      nil,
			repository:  mock.NewMockSummaryRepository(ctrl),
			expectedErr: "message reader is nil",
		},
		{
			name:        "nil repository",
			socials:     []client.Socials{mock.NewMockSocials(ctrl)},
			reader:      mock.NewMockMessageReader(ctrl),
			repository:  nil,
			expectedErr: "repository is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := service.NewPoster(tt.socials, tt.reader, tt.repository)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Nil(t, processor)
		})
	}
}

func TestPost_ValidContent_ShouldSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	social := mock.NewMockSocials(ctrl)
	socials := []client.Socials{social}
	reader := mock.NewMockMessageReader(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	poster, err := service.NewPoster(socials, reader, repo)
	require.NoError(t, err)
	require.NotNil(t, poster)
	defer poster.Close()

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)
			msgChan <- kgo.Record{Key: []byte("summary_id"), Value: []byte("test_id")}
			return nil
		},
	)
	reader.EXPECT().Close()
	repo.EXPECT().Get(gomock.Any(), "test_id").Return(&model.DailyFlightSummary{}, nil)
	social.EXPECT().PublishPost(gomock.Any(), gomock.Any()).Return(nil)

	err = poster.Post(context.Background())
	require.NoError(t, err)
}

func TestPost_RepoGetError_ShouldError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	social := mock.NewMockSocials(ctrl)
	socials := []client.Socials{social}
	reader := mock.NewMockMessageReader(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	poster, err := service.NewPoster(socials, reader, repo)
	require.NoError(t, err)
	require.NotNil(t, poster)
	defer poster.Close()

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)
			msgChan <- kgo.Record{Key: []byte("summary_id"), Value: []byte("test_id")}
			return nil
		},
	)
	reader.EXPECT().Close()
	repo.EXPECT().Get(gomock.Any(), "test_id").Return(nil, errors.New("test error"))

	err = poster.Post(context.Background())
	require.ErrorContains(t, err, "failed to get flight summary")
}

func TestPost_SocialPostError_ShouldError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	social := mock.NewMockSocials(ctrl)
	socials := []client.Socials{social}
	reader := mock.NewMockMessageReader(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	poster, err := service.NewPoster(socials, reader, repo)
	require.NoError(t, err)
	require.NotNil(t, poster)
	defer poster.Close()

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)
			msgChan <- kgo.Record{Key: []byte("summary_id"), Value: []byte("test_id")}
			return nil
		},
	)
	reader.EXPECT().Close()
	repo.EXPECT().Get(gomock.Any(), "test_id").Return(&model.DailyFlightSummary{}, nil)
	social.EXPECT().PublishPost(gomock.Any(), gomock.Any()).Return(errors.New("test error"))

	err = poster.Post(context.Background())
	require.ErrorContains(t, err, "failed to post content")
}

func TestPost_ContextCanceled_ShouldError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	social := mock.NewMockSocials(ctrl)
	socials := []client.Socials{social}
	reader := mock.NewMockMessageReader(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	poster, err := service.NewPoster(socials, reader, repo)
	require.NoError(t, err)
	require.NotNil(t, poster)
	defer poster.Close()

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case msgChan <- kgo.Record{Key: []byte("summary_id"), Value: []byte("test_id")}:
				cancel()
				time.Sleep(10 * time.Millisecond)
			}
			return nil
		},
	)
	reader.EXPECT().Close()
	repo.EXPECT().Get(gomock.Any(), "test_id").Return(&model.DailyFlightSummary{}, nil)
	social.EXPECT().PublishPost(gomock.Any(), gomock.Any()).Return(nil)

	err = poster.Post(ctx)
	require.ErrorContains(t, err, "context canceled while posting content")
}

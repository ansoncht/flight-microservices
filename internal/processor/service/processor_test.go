package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ansoncht/flight-microservices/internal/processor/model"
	"github.com/ansoncht/flight-microservices/internal/processor/repository"
	"github.com/ansoncht/flight-microservices/internal/processor/service"
	"github.com/ansoncht/flight-microservices/internal/test/mock"
	msgQueue "github.com/ansoncht/flight-microservices/pkg/kafka"
	msg "github.com/ansoncht/flight-microservices/pkg/model"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/mock/gomock"
)

func TestNewProcessor_ValidDependencies_ShouldSucceed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	writer := mock.NewMockMessageWriter(ctrl)
	reader := mock.NewMockMessageReader(ctrl)
	sum := mock.NewMockSummarizer(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	processor, err := service.NewProcessor(writer, reader, sum, repo)
	require.NoError(t, err)
	require.NotNil(t, processor)
	require.Equal(t, writer, processor.MessageWriter)
	require.Equal(t, reader, processor.MessageReader)
}

func TestNewProcessor_NilDependencies_ShouldReturnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		writer      msgQueue.MessageWriter
		reader      msgQueue.MessageReader
		summarizer  service.Summarizer
		repository  repository.SummaryRepository
		expectedErr string
	}{
		{
			name:        "nil writer",
			writer:      nil,
			reader:      mock.NewMockMessageReader(ctrl),
			summarizer:  mock.NewMockSummarizer(ctrl),
			repository:  mock.NewMockSummaryRepository(ctrl),
			expectedErr: "message writer is nil",
		},
		{
			name:        "nil reader",
			writer:      mock.NewMockMessageWriter(ctrl),
			reader:      nil,
			summarizer:  mock.NewMockSummarizer(ctrl),
			repository:  mock.NewMockSummaryRepository(ctrl),
			expectedErr: "message reader is nil",
		},
		{
			name:        "nil summarizer",
			writer:      mock.NewMockMessageWriter(ctrl),
			reader:      mock.NewMockMessageReader(ctrl),
			summarizer:  nil,
			repository:  mock.NewMockSummaryRepository(ctrl),
			expectedErr: "summarizer is nil",
		},
		{
			name:        "nil repository",
			writer:      mock.NewMockMessageWriter(ctrl),
			reader:      mock.NewMockMessageReader(ctrl),
			summarizer:  mock.NewMockSummarizer(ctrl),
			repository:  nil,
			expectedErr: "repository is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := service.NewProcessor(tt.writer, tt.reader, tt.summarizer, tt.repository)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Nil(t, processor)
		})
	}
}

func TestProcess_ValidMessage_ShouldSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	writer := mock.NewMockMessageWriter(ctrl)
	reader := mock.NewMockMessageReader(ctrl)
	sum := mock.NewMockSummarizer(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	processor, err := service.NewProcessor(writer, reader, sum, repo)
	require.NoError(t, err)
	require.NotNil(t, processor)

	flight1, err := json.Marshal(&msg.FlightRecord{Airline: "UA", FlightNumber: "123", Destination: "LAX"})
	require.NoError(t, err)
	require.NotNil(t, flight1)

	flight2, err := json.Marshal(&msg.FlightRecord{Airline: "AA", FlightNumber: "456", Destination: "SFO"})
	require.NoError(t, err)
	require.NotNil(t, flight2)

	messages := []kgo.Record{
		{Key: []byte("start_of_stream"), Value: []byte("JFK")},
		{Key: []byte("flight"), Value: flight1},
		{Key: []byte("flight"), Value: flight2},
		{Key: []byte("end_of_stream"), Value: []byte("2025-05-07")},
	}

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)
			for _, msg := range messages {
				msgChan <- msg
			}
			return nil
		},
	)

	expectedSummary := &model.DailyFlightSummary{
		Date:              model.ToMongoDateTime(time.Date(2025, 5, 7, 0, 0, 0, 0, time.UTC)),
		Airport:           "JFK",
		TotalFlights:      2,
		AirlineCounts:     map[string]int{"UA": 1, "AA": 1},
		DestinationCounts: map[string]int{"LAX": 1, "SFO": 1},
		TopDestinations:   []string{"LAX", "SFO"},
		TopAirlines:       []string{"UA", "AA"},
	}

	sum.EXPECT().SummarizeFlights(gomock.Any(), "2025-05-07", "JFK").Return(expectedSummary, nil)
	repo.EXPECT().Insert(gomock.Any(), *expectedSummary).Return("test_id", nil)
	writer.EXPECT().WriteMessage(gomock.Any(), []byte("summary_id"), []byte("test_id")).Return(nil)

	err = processor.Process(ctx)
	require.NoError(t, err)
}

func TestProcess_MalformedMessage_ShouldSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	writer := mock.NewMockMessageWriter(ctrl)
	reader := mock.NewMockMessageReader(ctrl)
	sum := mock.NewMockSummarizer(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	processor, err := service.NewProcessor(writer, reader, sum, repo)
	require.NoError(t, err)
	require.NotNil(t, processor)

	flight1, err := json.Marshal(&msg.FlightRecord{Airline: "UA", FlightNumber: "123", Destination: "LAX"})
	require.NoError(t, err)
	require.NotNil(t, flight1)

	messages := []kgo.Record{
		{Key: []byte("start_of_stream"), Value: []byte("JFK")},
		{Key: []byte("flight"), Value: flight1},
		{Key: []byte("flight"), Value: []byte("malformed")},
		{Key: []byte("end_of_stream"), Value: []byte("2025-05-07")},
	}

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)
			for _, msg := range messages {
				msgChan <- msg
			}
			return nil
		},
	)

	expectedSummary := &model.DailyFlightSummary{
		Date:              model.ToMongoDateTime(time.Date(2025, 5, 7, 0, 0, 0, 0, time.UTC)),
		Airport:           "JFK",
		TotalFlights:      2,
		AirlineCounts:     map[string]int{"UA": 1},
		DestinationCounts: map[string]int{"LAX": 1},
		TopDestinations:   []string{"LAX"},
		TopAirlines:       []string{"UA"},
	}

	sum.EXPECT().SummarizeFlights(gomock.Any(), "2025-05-07", "JFK").Return(expectedSummary, nil)
	repo.EXPECT().Insert(gomock.Any(), *expectedSummary).Return("test_id", nil)
	writer.EXPECT().WriteMessage(gomock.Any(), []byte("summary_id"), []byte("test_id")).Return(nil)

	err = processor.Process(ctx)
	require.NoError(t, err)
}

func TestProcess_ContextCanceledWhenRead_ShouldError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	writer := mock.NewMockMessageWriter(ctrl)
	reader := mock.NewMockMessageReader(ctrl)
	sum := mock.NewMockSummarizer(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	flight, err := json.Marshal(&msg.FlightRecord{Airline: "UA", FlightNumber: "123", Destination: "LAX"})
	require.NoError(t, err)
	require.NotNil(t, flight)

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case msgChan <- kgo.Record{Key: []byte("start_of_stream"), Value: []byte("JFK")}:
				cancel()
				time.Sleep(10 * time.Millisecond)
				return ctx.Err()
			}
		},
	)

	processor, err := service.NewProcessor(writer, reader, sum, repo)
	require.NoError(t, err)
	require.NotNil(t, processor)

	err = processor.Process(ctx)
	require.ErrorContains(t, err, "error while reading messages")
}

func TestProcess_ContextCanceled_ShouldError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	writer := mock.NewMockMessageWriter(ctrl)
	reader := mock.NewMockMessageReader(ctrl)
	sum := mock.NewMockSummarizer(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	flight, err := json.Marshal(&msg.FlightRecord{Airline: "UA", FlightNumber: "123", Destination: "LAX"})
	require.NoError(t, err)
	require.NotNil(t, flight)

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case msgChan <- kgo.Record{Key: []byte("start_of_stream"), Value: []byte("JFK")}:
				cancel()
				time.Sleep(10 * time.Millisecond)
			}
			return nil
		},
	)

	processor, err := service.NewProcessor(writer, reader, sum, repo)
	require.NoError(t, err)
	require.NotNil(t, processor)

	err = processor.Process(ctx)
	require.ErrorContains(t, err, "context canceled while processing messages")
}

func TestProcess_ReaderError_ShouldError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	writer := mock.NewMockMessageWriter(ctrl)
	reader := mock.NewMockMessageReader(ctrl)
	sum := mock.NewMockSummarizer(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	processor, err := service.NewProcessor(writer, reader, sum, repo)
	require.NoError(t, err)
	require.NotNil(t, processor)

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)
			return errors.New("test error")
		},
	)

	err = processor.Process(ctx)
	require.ErrorContains(t, err, "error while reading messages")
}

func TestProcess_SummarizeFlightsError_ShouldError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	writer := mock.NewMockMessageWriter(ctrl)
	reader := mock.NewMockMessageReader(ctrl)
	sum := mock.NewMockSummarizer(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	processor, err := service.NewProcessor(writer, reader, sum, repo)
	require.NoError(t, err)
	require.NotNil(t, processor)

	flight, err := json.Marshal(&msg.FlightRecord{Airline: "UA", FlightNumber: "123", Destination: "LAX"})
	require.NoError(t, err)
	require.NotNil(t, flight)

	messages := []kgo.Record{
		{Key: []byte("start_of_stream"), Value: []byte("JFK")},
		{Key: []byte("flight"), Value: flight},
		{Key: []byte("end_of_stream"), Value: []byte("2025-05-07")},
	}

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)
			for _, msg := range messages {
				msgChan <- msg
			}
			return nil
		},
	)

	sum.EXPECT().SummarizeFlights(gomock.Any(), "2025-05-07", "JFK").Return(nil, errors.New("test error"))

	err = processor.Process(ctx)
	require.ErrorContains(t, err, "failed to summarize flights")
}

func TestProcessor_Process_RepositoryInsertError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	writer := mock.NewMockMessageWriter(ctrl)
	reader := mock.NewMockMessageReader(ctrl)
	sum := mock.NewMockSummarizer(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	processor, err := service.NewProcessor(writer, reader, sum, repo)
	require.NoError(t, err)
	require.NotNil(t, processor)

	flight, err := json.Marshal(&msg.FlightRecord{Airline: "UA", FlightNumber: "123", Destination: "LAX"})
	require.NoError(t, err)
	require.NotNil(t, flight)

	messages := []kgo.Record{
		{Key: []byte("start_of_stream"), Value: []byte("JFK")},
		{Key: []byte("flight"), Value: flight},
		{Key: []byte("end_of_stream"), Value: []byte("2025-05-07")},
	}

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)
			for _, msg := range messages {
				msgChan <- msg
			}
			return nil
		},
	)

	expectedSummary := &model.DailyFlightSummary{
		Date:              model.ToMongoDateTime(time.Date(2025, 5, 7, 0, 0, 0, 0, time.UTC)),
		Airport:           "JFK",
		TotalFlights:      1,
		AirlineCounts:     map[string]int{"UA": 1},
		DestinationCounts: map[string]int{"LAX": 1},
		TopDestinations:   []string{"LAX"},
		TopAirlines:       []string{"UA"},
	}

	sum.EXPECT().SummarizeFlights(gomock.Any(), "2025-05-07", "JFK").Return(expectedSummary, nil)
	repo.EXPECT().Insert(gomock.Any(), *expectedSummary).Return("", errors.New("test error"))

	err = processor.Process(ctx)
	require.ErrorContains(t, err, "failed to insert summary")
}

func TestProcessor_Process_WriteMessageError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	writer := mock.NewMockMessageWriter(ctrl)
	reader := mock.NewMockMessageReader(ctrl)
	sum := mock.NewMockSummarizer(ctrl)
	repo := mock.NewMockSummaryRepository(ctrl)

	processor, err := service.NewProcessor(writer, reader, sum, repo)
	require.NoError(t, err)
	require.NotNil(t, processor)

	flight, err := json.Marshal(&msg.FlightRecord{Airline: "UA", FlightNumber: "123", Destination: "LAX"})
	require.NoError(t, err)
	require.NotNil(t, flight)

	messages := []kgo.Record{
		{Key: []byte("start_of_stream"), Value: []byte("JFK")},
		{Key: []byte("flight"), Value: flight},
		{Key: []byte("end_of_stream"), Value: []byte("2025-05-07")},
	}

	reader.EXPECT().ReadMessages(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, msgChan chan<- kgo.Record) error {
			defer close(msgChan)
			for _, msg := range messages {
				msgChan <- msg
			}
			return nil
		},
	)

	expectedSummary := &model.DailyFlightSummary{
		Date:              model.ToMongoDateTime(time.Date(2025, 5, 7, 0, 0, 0, 0, time.UTC)),
		Airport:           "JFK",
		TotalFlights:      1,
		AirlineCounts:     map[string]int{"UA": 1},
		DestinationCounts: map[string]int{"LAX": 1},
		TopDestinations:   []string{"LAX"},
		TopAirlines:       []string{"UA"},
	}

	sum.EXPECT().SummarizeFlights(gomock.Any(), "2025-05-07", "JFK").Return(expectedSummary, nil)
	repo.EXPECT().Insert(gomock.Any(), *expectedSummary).Return("test_id", nil)
	writer.EXPECT().WriteMessage(gomock.Any(), []byte("summary_id"), []byte("test_id")).Return(errors.New("test error"))

	err = processor.Process(ctx)
	require.ErrorContains(t, err, "failed to publish summary ObjectID")
}

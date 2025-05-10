package kafka_test

import (
	"context"
	"errors"
	"testing"
	"time"

	msgQueue "github.com/ansoncht/flight-microservices/pkg/kafka"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	testKafkaImage     = "confluentinc/confluent-local:7.5.0"
	testKafkaClusterID = "test-cluster"
	testTopic          = "test-topic"
	testGroupID        = "test-group"
	testMessageKey     = "test-key"
	testMessageValue   = "hello, world"
	timeout            = 10 * time.Second
)

// setupKafkaTest spins up a Kafka container for testing and returns its brokers and a cleanup function.
func setupKafkaTest(ctx context.Context, t *testing.T) (brokers []string, cleanup func()) {
	t.Helper()

	kafkaContainer, err := kafka.Run(ctx,
		testKafkaImage,
		kafka.WithClusterID(testKafkaClusterID),
	)
	require.NoError(t, err)
	require.NotNil(t, kafkaContainer)

	cleanup = func() {
		err := kafkaContainer.Terminate(ctx)
		require.NoError(t, err)
	}

	brokers, err = kafkaContainer.Brokers(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, brokers)

	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	require.NoError(t, err)
	defer client.Close()

	admin := kadm.NewClient(client)

	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	createTopicResponse, err := admin.CreateTopic(ctxTimeout, -1, -1, nil, testTopic)
	require.NoError(t, err)
	require.NotNil(t, createTopicResponse)

	return brokers, cleanup
}

func TestNewKafkaReader_ValidConfig_ShouldSucceed(t *testing.T) {
	cfg := msgQueue.ReaderConfig{
		Address: "localhost:9092",
		Topic:   testTopic,
		GroupID: testGroupID,
	}
	reader, err := msgQueue.NewKafkaReader(cfg)
	require.NoError(t, err)
	require.NotNil(t, reader)
}

func TestNewKafkaReader_InvalidConfig_ShouldError(t *testing.T) {
	cfg := msgQueue.ReaderConfig{
		Address: "localhost:9092",
		Topic:   testTopic,
		GroupID: testGroupID,
	}

	tests := []struct {
		name    string
		cfg     msgQueue.ReaderConfig
		wantErr string
	}{
		{
			name: "Missing Address",
			cfg: func() msgQueue.ReaderConfig {
				c := cfg
				c.Address = ""
				return c
			}(),
			wantErr: "kafka broker address is empty",
		},
		{
			name: "Missing Topic",
			cfg: func() msgQueue.ReaderConfig {
				c := cfg
				c.Topic = ""
				return c
			}(),
			wantErr: "kafka topic is empty",
		},
		{
			name: "Missing GroupID",
			cfg: func() msgQueue.ReaderConfig {
				c := cfg
				c.GroupID = ""
				return c
			}(),
			wantErr: "kafka group ID is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := msgQueue.NewKafkaReader(tt.cfg)
			require.ErrorContains(t, err, tt.wantErr)
			require.Nil(t, reader)
		})
	}
}

func TestReadMessages_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	brokers, cleanup := setupKafkaTest(ctx, t)
	defer cleanup()

	brokerAddress := brokers[0]

	wCfg := msgQueue.WriterConfig{
		Address: brokerAddress,
		Topic:   testTopic,
	}
	writer, err := msgQueue.NewKafkaWriter(wCfg)
	require.NoError(t, err)
	require.NotNil(t, writer)
	defer writer.Close()

	writeCtx, writeCancel := context.WithTimeout(ctx, 5*time.Second)
	defer writeCancel()

	err = writer.WriteMessage(writeCtx, []byte(testMessageKey), []byte(testMessageValue))
	require.NoError(t, err)

	rCfg := msgQueue.ReaderConfig{
		Address: brokerAddress,
		Topic:   testTopic,
		GroupID: testGroupID,
	}
	reader, err := msgQueue.NewKafkaReader(rCfg)
	require.NoError(t, err)
	require.NotNil(t, reader)
	defer reader.Close()

	t.Run("Successful ReadMessages", func(t *testing.T) {
		msgChan := make(chan kgo.Record, 1)
		readErrChan := make(chan error, 1)
		readCtx, readCancel := context.WithCancel(ctx)
		defer readCancel()

		go func() {
			readErrChan <- reader.ReadMessages(readCtx, msgChan)
		}()

		select {
		case msg := <-msgChan:
			require.Equal(t, testMessageKey, string(msg.Key))
			require.Equal(t, testMessageValue, string(msg.Value))
		case err := <-readErrChan:
			require.NoError(t, err)
		case <-time.After(timeout):
			t.Fatal("timed out waiting for message from Kafka")
		}

		readCancel()
		select {
		case err := <-readErrChan:
			if err != nil && !errors.Is(err, context.Canceled) {
				require.NoError(t, err)
			}
		case <-time.After(2 * time.Second):
			t.Error("ReadMessages goroutine did not exit cleanly after cancellation")
		}
	})

	t.Run("Nil Reader", func(t *testing.T) {
		msgChan := make(chan kgo.Record, 1)
		readCtx, readCancel := context.WithCancel(ctx)
		defer readCancel()

		var dummyReader *msgQueue.Reader
		err := dummyReader.ReadMessages(readCtx, msgChan)
		require.ErrorContains(t, err, "kafka reader is nil")
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()

		msgChan := make(chan kgo.Record)
		err = reader.ReadMessages(cancelCtx, msgChan)
		require.ErrorIs(t, err, context.Canceled)
	})
}

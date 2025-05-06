package kafka_test

import (
	"context"
	"testing"
	"time"

	msgQueue "github.com/ansoncht/flight-microservices/pkg/kafka"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

func TestNewKafkaWriter_ValidConfig_ShouldSucceed(t *testing.T) {
	cfg := msgQueue.WriterConfig{
		Address: "localhost:9092",
		Topic:   testTopic,
	}

	writer, err := msgQueue.NewKafkaWriter(cfg)
	require.NoError(t, err)
	require.NotNil(t, writer)
}

func TestNewKafkaWriter_InvalidConfig_ShouldError(t *testing.T) {
	cfg := msgQueue.WriterConfig{
		Address: "localhost:9092",
		Topic:   testTopic,
	}

	tests := []struct {
		name    string
		cfg     msgQueue.WriterConfig
		wantErr string
	}{
		{
			name: "Missing Address",
			cfg: func() msgQueue.WriterConfig {
				c := cfg
				c.Address = ""
				return c
			}(),
			wantErr: "kafka broker address is empty",
		},
		{
			name: "Missing Topic",
			cfg: func() msgQueue.WriterConfig {
				c := cfg
				c.Topic = ""
				return c
			}(),
			wantErr: "kafka topic is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := msgQueue.NewKafkaWriter(tt.cfg)
			require.ErrorContains(t, err, tt.wantErr)
			require.Nil(t, writer)
		})
	}
}

func TestWriteMessage_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	brokers, cleanup := setupKafkaTest(ctx, t)
	defer cleanup()

	brokerAddress := brokers[0]

	writer := &msgQueue.Writer{
		KafkaWriter: &kafkago.Writer{
			Addr:         kafkago.TCP(brokerAddress),
			Topic:        testTopic,
			RequiredAcks: kafkago.RequireOne,
			Async:        false,
			BatchTimeout: 100 * time.Millisecond,
		},
	}
	defer func() {
		err := writer.Close()
		require.NoError(t, err)
	}()

	t.Run("Successful WriteMessage", func(t *testing.T) {
		writeCtx, writeCancel := context.WithTimeout(ctx, 5*time.Second)
		defer writeCancel()

		err := writer.WriteMessage(writeCtx, []byte(testMessageKey), []byte(testMessageValue))
		require.NoError(t, err)

		cfg := msgQueue.ReaderConfig{
			Address: brokerAddress,
			Topic:   testTopic,
			GroupID: testGroupID + "-write-test",
		}

		reader, err := msgQueue.NewKafkaReader(cfg)
		require.NoError(t, err)
		require.NotNil(t, reader)
		defer func() {
			err := reader.Close()
			require.NoError(t, err)
		}()

		msgChan := make(chan kafkago.Message, 1)
		readErrChan := make(chan error, 1)
		readCtx, readCancel := context.WithTimeout(ctx, readTimeout)
		defer readCancel()

		go func() {
			readErrChan <- reader.ReadMessages(readCtx, msgChan)
		}()

		select {
		case msg := <-msgChan:
			require.Equal(t, []byte(testMessageKey), msg.Key)
			require.Equal(t, []byte(testMessageValue), msg.Value)
		case err := <-readErrChan:
			require.NoError(t, err)
		case <-readCtx.Done():
			t.Fatal("timed out waiting for message from Kafka after write")
		}
	})

	t.Run("Nil Writer", func(t *testing.T) {
		var dummyWriter *msgQueue.Writer
		err := dummyWriter.WriteMessage(ctx, []byte(testMessageKey), []byte(testMessageValue))
		require.ErrorContains(t, err, "kafka writer is nil")
	})

	t.Run("Empty Key", func(t *testing.T) {
		err := writer.WriteMessage(ctx, []byte{}, []byte(testMessageValue))
		require.ErrorContains(t, err, "message key is nil or empty")
	})

	t.Run("Empty Value", func(t *testing.T) {
		err := writer.WriteMessage(ctx, []byte(testMessageKey), []byte{})
		require.ErrorContains(t, err, "message value is nil or empty")
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()

		err := writer.WriteMessage(cancelCtx, []byte(testMessageKey), []byte(testMessageValue))
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestWriterClose_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	brokers, cleanup := setupKafkaTest(ctx, t)
	defer cleanup()

	t.Run("Successful Close", func(t *testing.T) {
		cfg := msgQueue.WriterConfig{
			Address: brokers[0],
			Topic:   testTopic,
		}

		writer, err := msgQueue.NewKafkaWriter(cfg)
		require.NoError(t, err)
		require.NotNil(t, writer)

		err = writer.Close()
		require.NoError(t, err)
	})

	t.Run("Nil Reader", func(t *testing.T) {
		var writer *msgQueue.Writer
		err := writer.Close()
		require.NoError(t, err)
	})
}

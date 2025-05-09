package mongo_test

import (
	"context"
	"testing"

	"github.com/ansoncht/flight-microservices/pkg/mongo"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func TestNewMongoClient_ValidConfig_ShouldSucceed(t *testing.T) {
	ctx := context.Background()

	// Start a MongoDB container
	mongodbContainer, err := mongodb.Run(ctx, "mongo:6")
	defer func() {
		err := testcontainers.TerminateContainer(mongodbContainer)
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	uri, err := mongodbContainer.ConnectionString(ctx)
	require.NoError(t, err)

	cfg := mongo.ClientConfig{
		URI:               uri,
		DB:                "testdb",
		PoolSize:          5,
		ConnectionTimeout: 10,
		SocketTimeout:     10,
	}
	mongo, err := mongo.NewMongoClient(ctx, cfg)
	defer func() {
		err = mongo.Client.Disconnect(ctx)
		require.NoError(t, err)
	}()
	require.NoError(t, err)
	require.NotNil(t, mongo)
	require.NotNil(t, mongo.Client)
	require.NotNil(t, mongo.Database)
	require.Equal(t, "testdb", mongo.Database.Name())
}

func TestNewMongoClient_InvalidConfig_ShouldError(t *testing.T) {
	ctx := context.Background()

	validCfg := mongo.ClientConfig{
		URI:               "mongodb://localhost:27017",
		DB:                "testdb",
		PoolSize:          5,
		ConnectionTimeout: 1,
		SocketTimeout:     1,
	}

	tests := []struct {
		name    string
		cfg     mongo.ClientConfig
		wantErr string
	}{
		{
			name: "Missing URI",
			cfg: func() mongo.ClientConfig {
				c := validCfg
				c.URI = ""
				return c
			}(),
			wantErr: "MongoDB connection URI is empty",
		},
		{
			name: "Invalid URI",
			cfg: func() mongo.ClientConfig {
				c := validCfg
				c.URI = "wrong_url"
				return c
			}(),
			wantErr: "failed to connect to MongoDB",
		},
		{
			name: "Missing DB",
			cfg: func() mongo.ClientConfig {
				c := validCfg
				c.DB = ""
				return c
			}(),
			wantErr: "MongoDB database name is empty",
		},
		{
			name: "Invalid Pool Size",
			cfg: func() mongo.ClientConfig {
				c := validCfg
				c.PoolSize = 0
				return c
			}(),
			wantErr: "MongoDB pool size is invalid",
		},
		{
			name: "Invalid Connection Timeout",
			cfg: func() mongo.ClientConfig {
				c := validCfg
				c.ConnectionTimeout = 0
				return c
			}(),
			wantErr: "MongoDB connection timeout is invalid",
		},
		{
			name: "Invalid Socket Timeout",
			cfg: func() mongo.ClientConfig {
				c := validCfg
				c.SocketTimeout = 0
				return c
			}(),
			wantErr: "MongoDB socket timeout is invalid",
		},
		{
			name: "Failed to Ping",
			cfg: func() mongo.ClientConfig {
				c := validCfg
				c.URI = "mongodb://localhost:27017"
				return c
			}(),
			wantErr: "failed to ping MongoDB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mongo, err := mongo.NewMongoClient(ctx, tt.cfg)
			require.ErrorContains(t, err, tt.wantErr)
			require.Nil(t, mongo)
		})
	}
}

package repository_test

import (
	"context"
	"testing"

	"github.com/ansoncht/flight-microservices/pkg/model"
	"github.com/ansoncht/flight-microservices/pkg/mongo"
	"github.com/ansoncht/flight-microservices/pkg/repository"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func TestNewMongoSummaryRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

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

	repo, err := repository.NewMongoSummaryRepository(mongo)
	require.NoError(t, err)
	require.NotNil(t, repo)
}

func TestNewMongoSummaryRepository_NilClient_ShouldError(t *testing.T) {
	var mongo *mongo.Client
	repo, err := repository.NewMongoSummaryRepository(mongo)
	require.ErrorContains(t, err, "mongo client is nil")
	require.Nil(t, repo)
}

func TestInsert_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

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

	repo, err := repository.NewMongoSummaryRepository(mongo)
	require.NoError(t, err)
	require.NotNil(t, repo)

	id, err := repo.Insert(ctx, model.DailyFlightSummary{})
	require.NoError(t, err)
	require.NotEmpty(t, id)
}

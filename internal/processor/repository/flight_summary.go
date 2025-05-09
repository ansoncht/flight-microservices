package repository

import (
	"context"
	"fmt"

	"github.com/ansoncht/flight-microservices/internal/processor/model"
	db "github.com/ansoncht/flight-microservices/pkg/mongo"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const dailySummaryCollection = "daily_summaries"

// SummaryRepository defines the interface for interacting with flight summary data in database.
type SummaryRepository interface {
	// Insert inserts a new flight summary into the database.
	Insert(ctx context.Context, summary model.DailyFlightSummary) (string, error)
}

// MongoSummaryRepository holds the MongoDB collection for flight summaries.
// It implements the SummaryRepository interface to provide methods for inserting flight summary data.
type MongoSummaryRepository struct {
	// Collection specifies the MongoDB collection for flight summaries.
	Collection *mongo.Collection
}

// NewMongoSummaryRepository creates a new MongoSummaryRepository instance based on the provided MongoDB collection.
func NewMongoSummaryRepository(client *db.Client) (*MongoSummaryRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("mongo client is nil")
	}

	collection := client.Database.Collection(dailySummaryCollection)

	return &MongoSummaryRepository{
		Collection: collection,
	}, nil
}

// Insert adds a flight summary to the MongoDB collection.
func (r *MongoSummaryRepository) Insert(ctx context.Context, summary model.DailyFlightSummary) (string, error) {
	result, err := r.Collection.InsertOne(ctx, summary)
	if err != nil {
		return "", fmt.Errorf("failed to insert to collection %s: %w", dailySummaryCollection, err)
	}

	// convert objectid to string
	oid, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("failed to cast InsertedID to ObjectID")
	}

	return oid.Hex(), nil
}

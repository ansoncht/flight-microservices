package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DailyFlightSummary holds aggregated statistics for all flights departing from a specific airport on a given day.
type DailyFlightSummary struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	Date              primitive.DateTime `bson:"date"`
	Airport           string             `bson:"airport"`
	TotalFlights      int                `bson:"totalFlights"`
	AirlineCounts     map[string]int     `bson:"airlineCounts"`
	DestinationCounts map[string]int     `bson:"destinationCounts"`
	TopDestinations   []string           `bson:"topDestinations,omitempty"`
	TopAirlines       []string           `bson:"topAirlines,omitempty"`
}

// ToMongoDateTime converts time.Time to primitive.DateTime for MongoDB.
func ToMongoDateTime(t time.Time) primitive.DateTime {
	return primitive.NewDateTimeFromTime(t)
}

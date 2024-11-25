package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// FlightData represents the data for a single flight.
type FlightData struct {
	Flight      string
	Origin      string
	Destination string
}

// FlightSummary represents a summary of flights for a specific date.
type FlightSummary struct {
	Date    primitive.DateTime `bson:"date"`
	Summary map[string]int     `bson:"summary"`
}

package model

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	limit = 5
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

// FormatForSocialMedia formats the DailyFlightSummary for social media content.
func (s *DailyFlightSummary) FormatForSocialMedia() string {
	// Convert MongoDB date to Go's time.Time
	date := s.Date.Time()

	// Limit the top airlines and destinations to 5
	topAirlines := s.TopAirlines
	if len(topAirlines) > limit {
		topAirlines = topAirlines[:limit]
	}

	topDestinations := s.TopDestinations
	if len(topDestinations) > limit {
		topDestinations = topDestinations[:limit]
	}

	// Format the summary with emojis
	return fmt.Sprintf(
		"✈️ **Daily Flight Summary** ✈️\n"+
			"📍 **Airport**: %s\n"+
			"📅 **Date**: %s\n"+
			"🛫 **Total Flights**: %d\n\n"+
			"🏆 **Top 5 Airlines**:\n%s\n\n"+
			"🌍 **Top 5 Destinations**:\n%s\n",
		s.Airport,
		date.Format("2006-01-02"), // Format date as YYYY-MM-DD
		s.TotalFlights,
		formatListWithNumbers(topAirlines),
		formatListWithNumbers(topDestinations),
	)
}

// formatListWithNumbers formats a list of strings with numbers (e.g., 1️⃣, 2️⃣).
func formatListWithNumbers(items []string) string {
	formatted := ""
	emojis := []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣"}
	for i, item := range items {
		if i < len(emojis) {
			formatted += fmt.Sprintf("%s %s\n", emojis[i], item)
		}
	}
	return formatted
}

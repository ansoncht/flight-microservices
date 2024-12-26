package model

// Flight represents a flight entry with essential details for gRPC requests.
type Flight struct {
	FlightNumber string // Flight number for the flight
	Airline      string // Airline operating the flight
	Origin       string // Departure airport of the flight
	Destination  string // Arrival airport of the flight.
	FirstSeen    int    // First timestamp when the flight was first detected
	LastSeen     int    // Last timestamp when the flight was last detected
}

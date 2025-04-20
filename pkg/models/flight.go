package models

// FlightRecord holds the essential details a flight entry.
type FlightRecord struct {
	FlightNumber string `json:"flightNumber"`
	Airline      string `json:"airline"`
	Origin       string `json:"origin"`
	Destination  string `json:"destination"`
	FirstSeen    int    `json:"firstSeen"`
	LastSeen     int    `json:"lastSeen"`
}

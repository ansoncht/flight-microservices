package model

// Flight represents a single flight entry.
type Flight struct {
	No      string `json:"no"`
	Airline string `json:"airline"`
}

// FlightList represents the list of flights for a specific time.
type FlightList struct {
	Time        string   `json:"time"`
	Flight      []Flight `json:"flight"`
	Status      string   `json:"status"`
	Destination []string `json:"destination"`
	Terminal    string   `json:"terminal"`
	Aisle       string   `json:"aisle"`
	Gate        string   `json:"gate"`
}

// FlightData represents the overall structure of the response.
type FlightData struct {
	Date    string       `json:"date"`
	Arrival bool         `json:"arrival"`
	Cargo   bool         `json:"cargo"`
	List    []FlightList `json:"list"`
}

// FlightDetail represents the flight detail being sent to server.
type FlightDetail struct {
	Flight      string
	Origin      string
	Destination string
}

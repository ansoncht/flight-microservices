package model

// FlightResponse represents the structure of flight data returned by the flight API.
type FlightResponse struct {
	Origin                 string `json:"estDepartureAirport"`              // Departure airport ICAO code
	Destination            string `json:"estArrivalAirport"`                // Arrival airport ICAO code
	CallSign               string `json:"callsign"`                         // Flight's call sign
	Icao24                 string `json:"icao24"`                           // Unique identifier for the aircraft
	FirstSeen              int    `json:"firstSeen"`                        // Timestamp when the flight was first detected
	LastSeen               int    `json:"lastSeen"`                         // Timestamp when the flight was last detected
	DepartureHorizDistance int    `json:"estDepartureAirportHorizDistance"` // Horizontal distance to the departure airport
	DepartureVertDistance  int    `json:"estDepartureAirportVertDistance"`  // Vertical distance to the departure airport
	ArrivalHorizDistance   int    `json:"estArrivalAirportHorizDistance"`   // Horizontal distance to the arrival airport
	ArrivalVertDistance    int    `json:"estArrivalAirportVertDistance"`    // Vertical distance to the arrival airport
	OritinAlt              int    `json:"departureAirportCandidatesCount"`  // Number of candidate departure airports
	DestinationAlt         int    `json:"arrivalAirportCandidatesCount"`    // Number of candidate arrival airports
}

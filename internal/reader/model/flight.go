package model

// Flight holds flight information from the external API.
type Flight struct {
	Origin                 string `json:"estDepartureAirport"`
	Destination            string `json:"estArrivalAirport"`
	Callsign               string `json:"callsign"`
	Icao24                 string `json:"icao24"`
	FirstSeen              int    `json:"firstSeen"`
	LastSeen               int    `json:"lastSeen"`
	DepartureHorizDistance int    `json:"estDepartureAirportHorizDistance"`
	DepartureVertDistance  int    `json:"estDepartureAirportVertDistance"`
	ArrivalHorizDistance   int    `json:"estArrivalAirportHorizDistance"`
	ArrivalVertDistance    int    `json:"estArrivalAirportVertDistance"`
	OritinAlt              int    `json:"departureAirportCandidatesCount"`
	DestinationAlt         int    `json:"arrivalAirportCandidatesCount"`
}

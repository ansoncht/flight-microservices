package model

// Route holds flight route information from the external API.
type Route struct {
	Response Response `json:"response"`
}

// Response holds the details of the flight route.
type Response struct {
	FlightRoute FlightRoute `json:"flightroute"`
}

// FlightRoute holds the details of a flight route.
type FlightRoute struct {
	CallSign     string  `json:"callsign"`
	CallSignICAO string  `json:"callsign_icao"`
	CallSignIATA string  `json:"callsign_iata"`
	Airline      Airline `json:"airline"`
	Origin       Airport `json:"origin"`
	Destination  Airport `json:"destination"`
}

// Airline holds the details of an airline.
type Airline struct {
	Name       string `json:"name"`
	ICAO       string `json:"icao"`
	IATA       string `json:"iata"`
	Country    string `json:"country"`
	CountryISO string `json:"country_iso"`
	CallSign   string `json:"callsign"`
}

// Airport holds the details of an airport.
type Airport struct {
	CountryISOName string  `json:"country_iso_name"`
	CountryName    string  `json:"country_name"`
	Elevation      int     `json:"elevation"`
	IATACode       string  `json:"iata_code"`
	ICAOCode       string  `json:"icao_code"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Municipality   string  `json:"municipality"`
	Name           string  `json:"name"`
}

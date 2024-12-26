package model

// RouteResponse represents the structure of route data returned by the route API.
type RouteResponse struct {
	Response Response `json:"response"` // The flight route information wrapped inside "response"
}

// FlightRoute represents the details of a flight route.
type FlightRoute struct {
	CallSign     string  `json:"callsign"`      // Flight's call sign (e.g., "CRK452")
	CallSignICAO string  `json:"callsign_icao"` // ICAO code for the flight (e.g., "CRK452")
	CallSignIATA string  `json:"callsign_iata"` // IATA code for the flight (e.g., "HX452")
	Airline      Airline `json:"airline"`       // Airline information (Airline struct)
	Origin       Airport `json:"origin"`        // Origin airport details (Airport struct)
	Destination  Airport `json:"destination"`   // Destination airport details (Airport struct)
}

// Airline represents the details of an airline.
type Airline struct {
	Name       string `json:"name"`        // Name of the airline (e.g., "Hong Kong Airlines")
	ICAO       string `json:"icao"`        // ICAO code of the airline (e.g., "CRK")
	IATA       string `json:"iata"`        // IATA code of the airline (e.g., "HX")
	Country    string `json:"country"`     // Country where the airline is based (e.g., "Hong Kong")
	CountryISO string `json:"country_iso"` // ISO code of the country (e.g., "HK")
	CallSign   string `json:"callsign"`    // Airline's call sign (e.g., "BAUHINIA")
}

// Airport represents the details of an airport.
type Airport struct {
	CountryISOName string  `json:"country_iso_name"` // ISO name of the country (e.g., "HK")
	CountryName    string  `json:"country_name"`     // Name of the country (e.g., "Hong Kong")
	Elevation      int     `json:"elevation"`        // Elevation of the airport in feet (e.g., 28)
	IATACode       string  `json:"iata_code"`        // IATA code of the airport (e.g., "HKG")
	ICAOCode       string  `json:"icao_code"`        // ICAO code of the airport (e.g., "VHHH")
	Latitude       float64 `json:"latitude"`         // Latitude of the airport (e.g., 22.308901)
	Longitude      float64 `json:"longitude"`        // Longitude of the airport (e.g., 113.915001)
	Municipality   string  `json:"municipality"`     // Municipality where the airport is located (e.g., "Hong Kong")
	Name           string  `json:"name"`             // Name of the airport (e.g., "Hong Kong International Airport")
}

// Response represents a single response.
type Response struct {
	FlightRoute FlightRoute `json:"flightroute"` // Flight route details
}

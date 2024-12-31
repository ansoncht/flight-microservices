package summarizer

type Summarizer struct {
	flightCounts map[string]int
}

func NewSummarizer() *Summarizer {
	return &Summarizer{
		flightCounts: make(map[string]int),
	}
}

func (s *Summarizer) AddFlight(destination string) {
	s.flightCounts[destination]++
}

func (s *Summarizer) GetFlightCounts() map[string]int {
	return s.flightCounts
}

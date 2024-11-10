package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/client"
)

func ServeHTTP(svr *http.Server) error {
	slog.Info("Starting http server on port " + svr.Addr)

	if err := svr.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func FetchHandler(fetcher client.FlightFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received request", "method", r.Method, "url", r.URL.String())

		if err := fetcher.FetchFlightsFromAPI(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Successfully triggered a manual fetch")
	}
}

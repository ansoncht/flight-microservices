package server

import (
	"fmt"
	"log/slog"
	"net/http"
)

func ServeHTTP(svr *http.Server) error {
	slog.Info("Starting http server on port " + svr.Addr)

	if err := svr.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func FetchHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received request", "method", r.Method, "url", r.URL.String())

	fmt.Fprintln(w, "Fetch endpoint hit!")
}

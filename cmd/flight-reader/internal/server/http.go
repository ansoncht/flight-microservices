package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/clients"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/config"
)

const hoursInADay = 24

type Server struct {
	httpServer *http.Server
	httpClient *clients.HTTPClient
	grpcClient *clients.GRPCClient
	URL        string
}

func NewHTTPServer(httpClient *clients.HTTPClient, grpcClient *clients.GRPCClient) (*Server, error) {
	slog.Info("Creating http server for the service")

	handlerCfg, err := config.MakeHTTPHandlerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get http handler config: %w", err)
	}

	srv := &Server{
		httpClient: httpClient,
		grpcClient: grpcClient,
		URL:        handlerCfg.URL,
	}

	// register endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/fetch", srv.fetchHandler())

	serverCfg, err := config.MakeHTTPServerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get http server config: %w", err)
	}

	if serverCfg.Port == "" {
		return nil, fmt.Errorf("empty port number")
	}

	srv.httpServer = &http.Server{
		Addr:              ":" + serverCfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(serverCfg.ReadHeaderTimeout) * time.Second,
	}

	return srv, nil
}

func (s *Server) ServeHTTP(ctx context.Context) error {
	slog.Info("Starting HTTP server", "port", s.httpServer.Addr)

	c := make(chan error)

	// Start the server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)

			c <- fmt.Errorf("failed to start http server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Stopping HTTP server due to context cancellation")

		return nil
	case err := <-c:
		return err
	}
}

func (s *Server) Close(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown HTTP server", "error", err)

		return fmt.Errorf("failed to shutdown http server: %w", err)
	}

	return nil
}

func (s *Server) ScheduleJob(ctx context.Context) error {
	slog.Info("Starting Scheduler")

	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if now.After(nextRun) {
		nextRun = nextRun.Add(hoursInADay * time.Hour)
	}

	durationUntilNextRun := nextRun.Sub(now)

	ticker := time.NewTicker(hoursInADay * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.processFlights(ctx); err != nil {
				return fmt.Errorf("failed to fetch flights: %w", err)
			}
			time.Sleep(durationUntilNextRun)
		case <-ctx.Done():
			slog.Info("Stopping scheduler due to context cancellation")

			return nil
		}
	}
}

func (s *Server) fetchHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

			return
		}

		slog.Info("Received request", "method", r.Method, "url", r.URL.String())

		if err := s.processFlights(r.Context()); err != nil {
			slog.Error("Failed to process flights", "error", err)

			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Successfully triggered a manual fetch")
	}
}

func (s *Server) processFlights(ctx context.Context) error {
	// fetch flight data
	flightData, err := s.httpClient.FetchFlightsFromAPI(ctx, s.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch flights: %w", err)
	}

	// send flight data to processor via gRPC
	if err := s.grpcClient.SendFlightStream(ctx, flightData); err != nil {
		return fmt.Errorf("failed to send flights to processor: %w", err)
	}

	return nil
}

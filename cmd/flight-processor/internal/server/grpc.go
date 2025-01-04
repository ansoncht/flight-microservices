package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/db"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/summarizer"
	pb "github.com/ansoncht/flight-microservices/proto/src/summarizer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GrpcServer represents the gRPC server structure.
type GrpcServer struct {
	pb.UnimplementedSummarizerServer
	server  *grpc.Server // gRPC server instance
	lis     net.Listener // Network listener for incoming connections
	mongoDB *db.Mongo    // MongoDB instance for database operations
}

// NewGRPC creates a new gRPC server instance.
func NewGRPC(mongoDB *db.Mongo) (*GrpcServer, error) {
	slog.Info("Creating gRPC server for the service")

	cfg, err := config.LoadGrpcServerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load gRPC server config: %w", err)
	}

	// Validate the port number
	if cfg.Port == "" {
		return nil, fmt.Errorf("empty port number")
	}

	port, _ := strconv.Atoi(cfg.Port)
	if port < 1 {
		return nil, fmt.Errorf("port number must be greater than 1")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	grpcServer := &GrpcServer{
		server:  s,
		lis:     lis,
		mongoDB: mongoDB,
	}

	pb.RegisterSummarizerServer(s, grpcServer)
	reflection.Register(s)

	return grpcServer, nil
}

// ServeGRPC starts the gRPC server and handles incoming requests.
func (g *GrpcServer) ServeGRPC(ctx context.Context) error {
	slog.Info("Starting gRPC server", "port", g.lis.Addr().String())

	c := make(chan error)

	// Start the server in a goroutine
	go func() {
		if err := g.server.Serve(g.lis); err != nil {
			slog.Error("Failed to start gRPC server", "error", err)
			c <- fmt.Errorf("failed to start gRPC server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Stopping gRPC server due to context cancellation")
		return nil
	case err := <-c:
		return err
	}
}

// Close gracefully shuts down the gRPC server.
func (g *GrpcServer) Close() {
	g.server.GracefulStop()
}

// PullFlight handles incoming flight data streams from clients.
func (g *GrpcServer) PullFlight(stream pb.Summarizer_PullFlightServer) error {
	slog.Info("Receiving stream of flight data from Flight Reader")

	transaction := int64(0)
	summarizer := summarizer.NewSummarizer()

	for {
		flight, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			slog.Info("Sent response to client", "total", transaction)

			if err := g.mongoDB.InsertSummary(stream.Context(), summarizer.GetFlightCounts(), "2024-11-08"); err != nil {
				return fmt.Errorf("failed to insert summary: %w", err)
			}

			if err := stream.SendAndClose(&pb.PullFlightResponse{
				Transaction: transaction,
			}); err != nil {
				return fmt.Errorf("failed to send response to client: %w", err)
			}

			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to pull flight data: %w", err)
		}

		transaction++
		summarizer.AddFlight(flight.Destination)

		slog.Debug("Processing flight sent by client",
			"flight", flight.Flight,
			"origin", flight.Origin,
			"destination", flight.Destination,
		)
	}
}

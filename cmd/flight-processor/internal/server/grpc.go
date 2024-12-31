package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/db"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/summarizer"
	pb "github.com/ansoncht/flight-microservices/proto/src/summarizer"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	pb.UnimplementedSummarizerServer
	server  *grpc.Server
	lis     net.Listener
	mongoDB *db.Mongo
}

func NewGRPC(mongoDB *db.Mongo) (*GRPCServer, error) {
	slog.Info("Creating gRPC server for the service")

	cfg, err := config.LoadGrpcServerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get gRPC server config: %w", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	grpcServer := &GRPCServer{
		server:  s,
		lis:     lis,
		mongoDB: mongoDB,
	}

	pb.RegisterSummarizerServer(s, grpcServer)

	return grpcServer, nil
}

func (g *GRPCServer) ServeGRPC(ctx context.Context) error {
	slog.Info("Starting gRPC server", "port", g.lis.Addr().String())

	c := make(chan error)

	// Start the server in a goroutine
	go func() {
		if err := g.server.Serve(g.lis); err != nil {
			slog.Error("gRPC server error", "error", err)
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

func (g *GRPCServer) Close() {
	g.server.GracefulStop()
}

func (g *GRPCServer) PullFlight(stream pb.Summarizer_PullFlightServer) error {
	slog.Info("Receiving stream of flight data from reader")

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

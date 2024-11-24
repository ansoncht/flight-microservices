package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/config"
	pb "github.com/ansoncht/flight-microservices/proto/src/summarizer"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	server *grpc.Server
	lis    net.Listener
}

type server struct {
	pb.SummarizerServer
}

func NewGRPCServer() (*GRPCServer, error) {
	slog.Info("Creating gRPC server for the service")

	cfg, err := config.MakeGRPCServerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get gRPC server config: %w", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.PORT))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	pb.RegisterSummarizerServer(s, &server{})

	return &GRPCServer{
		server: s,
		lis:    lis,
	}, nil
}

func (s *GRPCServer) ServeGRPC(ctx context.Context) error {
	slog.Info("Starting gRPC server", "port", s.lis.Addr().String())

	c := make(chan error)

	// Start the server in a goroutine
	go func() {
		if err := s.server.Serve(s.lis); err != nil {
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

func (s *GRPCServer) Close() {
	s.server.GracefulStop()
}

func (s *server) PullFlight(stream pb.Summarizer_PullFlightServer) error {
	slog.Info("Receiving stream of flight data from reader")
	transaction := int64(0)

	for {
		flight, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			slog.Info("Sent response to client", "total", transaction)

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

		slog.Debug("flight info",
			"flight no", flight.Flight,
			"origin", flight.Origin,
			"destination", flight.Destination,
		)
	}
}

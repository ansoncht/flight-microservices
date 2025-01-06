package client

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/model"
	pb "github.com/ansoncht/flight-microservices/proto/src/summarizer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	client pb.SummarizerClient
	conn   *grpc.ClientConn
	stream grpc.ClientStreamingClient[pb.PullFlightRequest, pb.PullFlightResponse]
}

func NewGRPC() (*GrpcClient, error) {
	slog.Info("Creating gRPC client for the service")

	cfg, err := config.LoadGrpcClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load gRPC client config: %w", err)
	}

	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return &GrpcClient{
		client: pb.NewSummarizerClient(conn),
		conn:   conn,
		stream: nil,
	}, nil
}

func (g *GrpcClient) Close() error {
	if err := g.conn.Close(); err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}

	return nil
}

func (g *GrpcClient) StartStream(ctx context.Context) error {
	slog.Info("Starting stream to send flight data to processor")

	stream, err := g.client.PullFlight(ctx)
	if err != nil {
		return fmt.Errorf("failed to start gRPC stream: %w", err)
	}

	g.stream = stream

	return nil
}

func (g *GrpcClient) SendFlight(flight *model.Flight) error {
	slog.Info("Sending single flight detail to processor")

	slog.Debug("Sending flight to server",
		"flight", flight.FlightNumber,
		"origin", flight.Origin,
		"destination", flight.Destination,
	)

	if err := g.stream.Send(&pb.PullFlightRequest{
		Flight:      flight.FlightNumber,
		Origin:      flight.Origin,
		Destination: flight.Destination,
	}); err != nil {
		return fmt.Errorf("failed to send flight: %w", err)
	}

	return nil
}

func (g *GrpcClient) CloseStream() error {
	_, err := g.stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to close gRPC stream: %w", err)
	}

	g.stream = nil

	return nil
}

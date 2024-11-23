package clients

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-reader/internal/model"
	pb "github.com/ansoncht/flight-microservices/proto/src/summarizer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.SummarizerClient
}

func NewGRPCClient() (*GRPCClient, error) {
	slog.Info("Creating gRPC client for the service")

	cfg, err := config.MakeGRPCClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get gRPC client config: %w", err)
	}

	conn, err := grpc.NewClient(cfg.ADDRESS, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return &GRPCClient{
		conn:   conn,
		client: pb.NewSummarizerClient(conn),
	}, nil
}

func (c *GRPCClient) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}

	return nil
}

func (c *GRPCClient) SendFlightStream(
	ctx context.Context,
	flightData []model.FlightData,
) error {
	slog.Info("Sending stream of flight data to processor")

	stream, err := c.client.PullFlight(ctx)
	if err != nil {
		return fmt.Errorf("failed to create gRPC stream: %w", err)
	}

	for _, data := range flightData {
		for _, flightList := range data.List {
			// get the last destination
			var destination string
			if len(flightList.Destination) > 0 {
				destination = flightList.Destination[len(flightList.Destination)-1]
			}

			// get the flying flight number
			var flightNo string
			if len(flightList.Flight) > 0 {
				flightNo = flightList.Flight[0].No
			}

			slog.Debug("Sending flight to server",
				"flight", flightNo,
				"origin", "HKG",
				"destination", destination,
			)

			if err := stream.Send(&pb.PullFlightRequest{
				Flight:      flightNo,
				Origin:      "HKG",
				Destination: destination,
			}); err != nil {
				return fmt.Errorf("failed to send stream: %w", err)
			}
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to close stream: %w", err)
	}

	slog.Info("Received response from server", "total", strconv.Itoa(int(resp.Transaction)))

	return nil
}

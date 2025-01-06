package client

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/config"
	pb "github.com/ansoncht/flight-microservices/proto/src/poster"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	client pb.PosterClient
	conn   *grpc.ClientConn
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
		client: pb.NewPosterClient(conn),
		conn:   conn,
	}, nil
}

func (g *GrpcClient) Close() error {
	if err := g.conn.Close(); err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}

	return nil
}

func (g *GrpcClient) SendSummary(ctx context.Context, flightData map[string]int, date string) error {
	destinations := make([]string, 0, len(flightData))
	for destination := range flightData {
		destinations = append(destinations, destination)
	}

	sort.SliceStable(destinations, func(i, j int) bool {
		return flightData[destinations[i]] > flightData[destinations[j]]
	})

	var stats []*pb.FlightStat
	for _, destination := range destinations[:20] {
		flightStat := &pb.FlightStat{
			Destination: destination,
			Frequency:   int64(flightData[destination]),
		}
		stats = append(stats, flightStat)
	}

	req := &pb.SendSummaryRequest{
		Date:        date,
		FlightStats: stats,
	}

	_, err := g.client.SendSummary(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send request to flight poster: %w", err)
	}
	return nil
}

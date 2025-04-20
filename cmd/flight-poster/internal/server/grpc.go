package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-poster/internal/poster"
	pb "github.com/ansoncht/flight-microservices/proto/src/poster"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GrpcServer represents the gRPC server structure.
type GrpcServer struct {
	pb.UnimplementedPosterServer
	server  *grpc.Server    // gRPC server instance
	lis     net.Listener    // Network listener for incoming connections
	posters []poster.Poster // Posters for social media
}

// NewGRPC creates a new gRPC server instance.
func NewGRPC(cfg config.GrpcServerConfig, posters []poster.Poster) (*GrpcServer, error) {
	slog.Info("Creating gRPC server for the service")

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
		posters: posters,
	}

	pb.RegisterPosterServer(s, grpcServer)
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

func (g *GrpcServer) SendSummary(ctx context.Context, req *pb.SendSummaryRequest) (*pb.SendSummaryResponse, error) {
	slog.Info("Receiving flight summary from Flight Processor, posting to social media")

	// Create a message for social media
	message := fmt.Sprintf("✈️ Flight Summary for %s:\n", req.Date)

	// Iterate over the flight statistics to append all destinations
	for _, flightStat := range req.FlightStats {
		message += fmt.Sprintf("🌍 %s: %d\t\t", flightStat.Destination, flightStat.Frequency)
	}

	for _, poster := range g.posters {
		_, err := poster.PublishPost(ctx, message)
		if err != nil {
			return nil, fmt.Errorf("failed to post: %w", err)
		}
	}

	return &pb.SendSummaryResponse{}, nil
}

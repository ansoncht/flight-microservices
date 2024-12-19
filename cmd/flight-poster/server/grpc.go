package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/ansoncht/flight-microservices/cmd/flight-poster/config"
	pb "github.com/ansoncht/flight-microservices/proto/src/poster"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	server *grpc.Server
	lis    net.Listener
}

type server struct {
	pb.UnimplementedPosterServer
}

func NewGRPCServer() (*GrpcServer, error) {
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
	pb.RegisterPosterServer(s, &server{})

	return &GrpcServer{
		server: s,
		lis:    lis,
	}, nil
}

func (s *GrpcServer) ServeGRPC(ctx context.Context) error {
	slog.Info("Starting gRPC server", "port", s.lis.Addr().String())

	c := make(chan error)

	// Start the server in a goroutine
	go func() {
		if err := s.server.Serve(s.lis); err != nil {
			slog.Error("gRPC server error", "error", err)

			c <- err
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

func (s *GrpcServer) Close() {
	slog.Info("Stopping gRPC server gracefully")

	s.server.GracefulStop()
}

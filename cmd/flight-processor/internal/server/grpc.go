package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/client"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/config"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/db"
	"github.com/ansoncht/flight-microservices/cmd/flight-processor/internal/summarizer"
	pb "github.com/ansoncht/flight-microservices/proto/src/summarizer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GrpcServer holds the gRPC server and its configuration settings.
type GrpcServer struct {
	// UnimplementedSummarizerServer is the default implementation
	// for the gRPC Summarizer server interface.
	pb.UnimplementedSummarizerServer
	// server is the instance of the gRPC server handling incoming requests.
	server *grpc.Server
	// lis is the network listener for accepting incoming gRPC connections.
	lis net.Listener
	// mongoDB is the MongoDB instance used for database operations.
	mongoDB *db.Mongo
	// grpcClient is the gRPC client used to send summaries to another service.
	grpcClient *client.GrpcClient
}

// NewGRPC creates a new gRPC server based on the provided configuration..
func NewGRPC(cfg config.GrpcServerConfig, mongoDB *db.Mongo, grpcClient *client.GrpcClient) (*GrpcServer, error) {
	slog.Info("Creating gRPC server for the service")

	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 {
		return nil, fmt.Errorf("invalid port number: %s", cfg.Port)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	grpcServer := &GrpcServer{
		server:     grpc.NewServer(),
		lis:        lis,
		mongoDB:    mongoDB,
		grpcClient: grpcClient,
	}

	pb.RegisterSummarizerServer(grpcServer.server, grpcServer)
	reflection.Register(grpcServer.server)

	return grpcServer, nil
}

// ServeGRPC starts the gRPC server and handles incoming requests.
func (g *GrpcServer) ServeGRPC(ctx context.Context) error {
	slog.Info("Starting gRPC server", "port", g.lis.Addr().String())

	c := make(chan error)

	// Start the server in a goroutine
	go func() {
		if err := g.server.Serve(g.lis); err != nil {
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

// Close shuts down the gRPC server gracefully.
func (g *GrpcServer) Close() {
	g.server.GracefulStop()
}

// PullFlight handles incoming flight data streams from clients.
func (g *GrpcServer) PullFlight(stream pb.Summarizer_PullFlightServer) error {
	slog.Info("Receiving stream of flight data from Flight Reader")

	transaction := 0
	summarizer := summarizer.NewSummarizer()

	for {
		flight, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			slog.Info("Received all flights from Flight Reader", "total", transaction)

			ctx := stream.Context()
			flightData := summarizer.GetFlightCounts()
			date := time.Now().Format("2006-01-02")

			if err := g.mongoDB.InsertSummary(ctx, flightData, date); err != nil {
				return fmt.Errorf("failed to insert summary into MongoDB: %w", err)
			}

			if err := g.grpcClient.SendSummary(ctx, flightData, date); err != nil {
				return fmt.Errorf("failed to send summary to flight poster: %w", err)
			}

			if err := stream.SendAndClose(&pb.PullFlightResponse{}); err != nil {
				return fmt.Errorf("failed to send response to flight reader: %w", err)
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

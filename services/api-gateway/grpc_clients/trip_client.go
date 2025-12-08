package grpcclients

import (
	"os"

	pb "github.com/tenteedee/mini-uber/shared/proto/trip"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TripServiceClient struct {
	Client pb.TripServiceClient
	conn   *grpc.ClientConn
}

func NewTripServiceClient() (*TripServiceClient, error) {
	tripServiceURL := os.Getenv("TRIP_SERVICE_URL")
	if tripServiceURL == "" {
		tripServiceURL = "trip-service:9093"
	}

	conn, err := grpc.NewClient(tripServiceURL, grpc.WithTransportCredentials(
		insecure.NewCredentials(),
	))
	if err != nil {
		return nil, err
	}

	client := pb.NewTripServiceClient(conn)

	return &TripServiceClient{
		Client: client,
		conn:   conn,
	}, nil
}

func (c *TripServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

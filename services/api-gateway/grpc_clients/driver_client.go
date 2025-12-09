package grpcclients

import (
	"os"

	pb "github.com/tenteedee/mini-uber/shared/proto/driver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DriverServiceClient struct {
	Client pb.DriverServiceClient
	conn   *grpc.ClientConn
}

func NewDriverServiceClient() (*DriverServiceClient, error) {
	driverServiceURL := os.Getenv("DRIVER_SERVICE_URL")
	if driverServiceURL == "" {
		driverServiceURL = "driver-service:9092"
	}

	conn, err := grpc.NewClient(driverServiceURL, grpc.WithTransportCredentials(
		insecure.NewCredentials(),
	))
	if err != nil {
		return nil, err
	}

	client := pb.NewDriverServiceClient(conn)

	return &DriverServiceClient{
		Client: client,
		conn:   conn,
	}, nil
}

func (c *DriverServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

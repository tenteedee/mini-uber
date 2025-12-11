package domain

import pb "github.com/tenteedee/mini-uber/shared/proto/driver"

type DriverInMap struct {
	Driver *pb.Driver
}

type DriverService interface {
	RegisterDriver(driverId string, packageSlug string) (*pb.Driver, error)
	UnregisterDriver(driverId string)
	FindAvailableDrivers(packageType string) []string
}

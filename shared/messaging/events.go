package messaging

import (
	pb "github.com/tenteedee/mini-uber/shared/proto/trip"
)

const (
	FindAvailableDriversQueue = "find_available_drivers"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}

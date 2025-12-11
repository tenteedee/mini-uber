package messaging

import (
	pbd "github.com/tenteedee/mini-uber/shared/proto/driver"
	pb "github.com/tenteedee/mini-uber/shared/proto/trip"
)

const (
	FindAvailableDriversQueue        = "find_available_drivers"
	DriverCmdTripRequestQueue        = "driver_cmd_trip_request"
	DriverTripResponseQueue          = "driver_trip_response"
	NotifyDriversNoDriversFoundQueue = "notify_drivers_no_drivers_found"
	NotifyDriverAssignQueue          = "notify_driver_assign"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}

type DriverTripResponseData struct {
	Driver  *pbd.Driver `json:"driver"`
	TripID  string      `json:"trip_id"`
	RiderId string      `json:"rider_id"`
}

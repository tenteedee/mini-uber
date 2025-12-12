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
	PaymentTripResponseQueue         = "payment_trip_response"
	NotifyPaymentSessionCreatedQueue = "notify_payment_session_created"
	NotifyPaymentSuccessQueue        = "payment_success"
)

const DeadLetterQueue = "dead_letter_queue"

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}

type DriverTripResponseData struct {
	Driver  *pbd.Driver `json:"driver"`
	TripId  string      `json:"tripId"`
	RiderId string      `json:"riderId"`
}

type PaymentEventSessionCreatedData struct {
	TripID    string  `json:"tripId"`
	SessionID string  `json:"sessionId"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

type PaymentTripResponseData struct {
	TripID   string  `json:"tripId"`
	UserID   string  `json:"userId"`
	DriverID string  `json:"driverId"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type PaymentStatusUpdateData struct {
	TripID   string `json:"tripId"`
	UserID   string `json:"userId"`
	DriverID string `json:"driverId"`
}

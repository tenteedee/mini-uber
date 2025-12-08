package main

import (
	pb "github.com/tenteedee/mini-uber/shared/proto/trip"
	"github.com/tenteedee/mini-uber/shared/types"
)

type previewTripRequest struct {
	UserId      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (p *previewTripRequest) ToProto() *pb.PreviewTripRequest {
	return &pb.PreviewTripRequest{
		UserID: p.UserId,
		Pickup: &pb.Coordinate{
			Latitude:  p.Pickup.Latitude,
			Longitude: p.Pickup.Longitude,
		},
		Destination: &pb.Coordinate{
			Latitude:  p.Destination.Latitude,
			Longitude: p.Destination.Longitude,
		},
	}
}

type startTripRequest struct {
	RideFareID string `json:"rideFareID"`
	UserID     string `json:"userID"`
}

func (c *startTripRequest) toProto() *pb.CreateTripRequest {
	return &pb.CreateTripRequest{
		RideFareID: c.RideFareID,
		UserID:     c.UserID,
	}
}

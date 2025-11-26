package domain

import (
	"context"

	"github.com/tenteedee/mini-uber/shared/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripModel struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   string
	Status   string
	RideFare *RideFareModel
}

type TripRepository interface {
	CreateTrip(ctx context.Context, trip TripModel) (*TripModel, error)
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup *types.Coordinate, destination *types.Coordinate) (*types.OsrmApiResponse, error)
}

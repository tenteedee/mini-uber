package domain

import (
	tripTypes "github.com/tenteedee/mini-uber/services/trip-service/pkg/types"
	pb "github.com/tenteedee/mini-uber/shared/proto/trip"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserID            string
	PackageSlug       string // van, suv, sedan
	TotalPriceInCents float64
	ExpiresAt         string
	Route             *tripTypes.OsrmApiResponse
}

func (r *RideFareModel) ToProto() *pb.Ridefare {
	return &pb.Ridefare{
		Id:                r.ID.Hex(),
		UserID:            r.UserID,
		PackageSlug:       r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*pb.Ridefare {
	protoFares := make([]*pb.Ridefare, len(fares))

	for i, fare := range fares {
		protoFares[i] = fare.ToProto()
	}

	return protoFares
}

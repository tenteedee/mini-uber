package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type RideFareModel struct {
	ID          primitive.ObjectID
	UserId      string
	PackageSlug string // van, limo, sedan
	TotalPrice  float64
	ExpiresAt   string
}

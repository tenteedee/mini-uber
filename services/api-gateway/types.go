package main

import "github.com/tenteedee/mini-uber/shared/types"

type previewTripRequest struct {
	UserId      string           `json:"user_id"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

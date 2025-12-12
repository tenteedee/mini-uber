package types

import pb "github.com/tenteedee/mini-uber/shared/proto/trip"

// type OsrmApiResponse struct {
// 	Routes []struct {
// 		Distance float64 `json:"distance"`
// 		Duration float64 `json:"duration"`
// 		Geometry struct {
// 			Coordinates [][]float64 `json:"coordinates"`
// 		} `json:"geometry"`
// 	}
// }

type OsrmApiResponse struct {
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
			Type        string      `json:"type"`
		} `json:"geometry"`
	} `json:"routes"`

	Code string `json:"code"`
}

func (o *OsrmApiResponse) ToProto() *pb.Route {
	route := o.Routes[0]
	geometry := route.Geometry.Coordinates
	coordinates := make([]*pb.Coordinate, len(geometry))

	for i, coord := range geometry {
		coordinates[i] = &pb.Coordinate{
			Latitude:  coord[1],
			Longitude: coord[0],
		}
	}

	return &pb.Route{
		Distance: route.Distance,
		Duration: route.Duration,
		Geometry: []*pb.Geometry{
			{Coordinates: coordinates},
		},
	}
}

type PricingConfig struct {
	PricePerUnitOfDistance float64
	PricePerMinute         float64
}

func DefaultPricingConfig() *PricingConfig {
	return &PricingConfig{
		PricePerUnitOfDistance: 1.0,
		PricePerMinute:         0.25,
	}
}

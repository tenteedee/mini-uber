package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tenteedee/mini-uber/services/trip-service/internal/domain"
	"github.com/tenteedee/mini-uber/shared/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type service struct {
	repo domain.TripRepository
}

func NewService(repo domain.TripRepository) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	trip := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserId,
		Status:   "pending",
		RideFare: fare,
	}
	return s.repo.CreateTrip(ctx, *trip)

}

func (s *service) GetRoute(ctx context.Context, pickup *types.Coordinate, destination *types.Coordinate) (*types.OsrmApiResponse, error) {
	url := fmt.Sprintf(
		"http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		pickup.Longitude, pickup.Latitude,
		destination.Longitude, destination.Latitude,
	)

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route from OSRM API: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OSRM API response body: %v", err)
	}

	var routeResponse types.OsrmApiResponse

	if err := json.Unmarshal(body, &routeResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OSRM API response: %v", err)
	}

	return nil, nil
}

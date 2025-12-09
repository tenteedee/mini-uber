package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tenteedee/mini-uber/services/trip-service/internal/domain"
	tripTypes "github.com/tenteedee/mini-uber/services/trip-service/pkg/types"
	"github.com/tenteedee/mini-uber/shared/proto/trip"
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
		UserID:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
		Driver:   &trip.TripDriver{},
	}
	return s.repo.CreateTrip(ctx, trip)

}

func (s *service) GetTripRoute(ctx context.Context, pickup *types.Coordinate, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error) {
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

	var routeResponse tripTypes.OsrmApiResponse

	if err := json.Unmarshal(body, &routeResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OSRM API response: %v", err)
	}

	return &routeResponse, nil
}

func (s *service) EstimatePackagesPriceWithRoutes(route *tripTypes.OsrmApiResponse) []*domain.RideFareModel {
	baseFares := getBaseFares()
	estimatedFares := make([]*domain.RideFareModel, len(baseFares))

	for i, fare := range baseFares {
		estimatedFares[i] = estimateFareRoute(fare, route)
	}

	return estimatedFares
}

func estimateFareRoute(fare *domain.RideFareModel, route *tripTypes.OsrmApiResponse) *domain.RideFareModel {
	pricingCfg := tripTypes.DefaultPricingConfig()
	carPackagePrice := fare.TotalPriceInCents

	distance := route.Routes[0].Distance
	duration := route.Routes[0].Duration

	// distance
	distanceFare := distance * pricingCfg.PricePerUnitOfDistance
	//time
	timeFare := duration * pricingCfg.PricePerMinute
	// car price
	totalPrice := carPackagePrice + distanceFare + timeFare

	return &domain.RideFareModel{
		PackageSlug:       fare.PackageSlug,
		TotalPriceInCents: totalPrice,
	}
}

func (s *service) GenerateTripFares(ctx context.Context, rideFares []*domain.RideFareModel, userId string, route *tripTypes.OsrmApiResponse) ([]*domain.RideFareModel, error) {
	fares := make([]*domain.RideFareModel, len(rideFares))

	for i, fare := range rideFares {
		fare := &domain.RideFareModel{
			ID:                primitive.NewObjectID(),
			UserID:            userId,
			PackageSlug:       fare.PackageSlug,
			TotalPriceInCents: fare.TotalPriceInCents,
			Route:             route,
		}
		if err := s.repo.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("failed to save ride fare: %v", err)
		}

		fares[i] = fare
	}

	return fares, nil
}

func (s *service) GetAndValidateFare(ctx context.Context, fareID string, userID string) (*domain.RideFareModel, error) {
	fare, err := s.repo.GetRideFareByID(ctx, fareID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride fare by ID: %v", err)
	}

	if fare.UserID != userID {
		return nil, fmt.Errorf("fare does not belong to user")
	}
	return fare, nil
}

func getBaseFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{
			PackageSlug:       "sedan",
			TotalPriceInCents: 200,
		},
		{
			PackageSlug:       "suv",
			TotalPriceInCents: 300,
		},
		{
			PackageSlug:       "luxury",
			TotalPriceInCents: 1000,
		},
		{
			PackageSlug:       "van",
			TotalPriceInCents: 400,
		},
	}
}

package repository

import (
	"context"
	"fmt"

	"github.com/tenteedee/mini-uber/services/trip-service/internal/domain"
)

type inmemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInmemRepository() *inmemRepository {
	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inmemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.trips[trip.ID.Hex()] = trip
	return trip, nil
}

func (r *inmemRepository) SaveRideFare(ctx context.Context, fare *domain.RideFareModel) error {
	r.rideFares[fare.ID.Hex()] = fare
	return nil
}

func (r *inmemRepository) GetRideFareByID(ctx context.Context, fareID string) (*domain.RideFareModel, error) {
	fare, exists := r.rideFares[fareID]
	if !exists {
		return nil, fmt.Errorf("fare does not exists with ID: %s", fareID)
	}
	return fare, nil
}

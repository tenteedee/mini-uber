package main

import (
	"context"
	"log"

	"github.com/tenteedee/mini-uber/services/trip-service/internal/domain"
	"github.com/tenteedee/mini-uber/services/trip-service/internal/infrastructure/repository"
	"github.com/tenteedee/mini-uber/services/trip-service/internal/service"
)

func main() {
	ctx := context.Background()

	inmemRepo := repository.NewInmemRepository()
	service := service.NewService(inmemRepo)

	fare := &domain.RideFareModel{
		UserId: "12",
	}

	trip, err := service.CreateTrip(ctx, fare)
	if err != nil {
		log.Println(err)
	}

	log.Println(trip)
}

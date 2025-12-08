package grpc

import (
	"context"
	"log"

	"github.com/tenteedee/mini-uber/services/trip-service/internal/domain"
	pb "github.com/tenteedee/mini-uber/shared/proto/trip"
	"github.com/tenteedee/mini-uber/shared/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer
	service domain.TripService
}

func NewgRPCHandler(server *grpc.Server, service domain.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	log.Printf("PreviewTrip called: pickup=%v dest=%v", req.GetPickup(), req.GetDestination())

	pickup := req.GetPickup()
	destination := req.GetDestination()

	pickupCoordinates := &types.Coordinate{
		Latitude:  pickup.GetLatitude(),
		Longitude: pickup.GetLongitude(),
	}
	destinationCoordinates := &types.Coordinate{
		Latitude:  destination.GetLatitude(),
		Longitude: destination.GetLongitude(),
	}

	route, err := h.service.GetTripRoute(ctx,
		pickupCoordinates,
		destinationCoordinates,
	)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route %v: ", err)
	}

	userID := req.GetUserID()

	// estimate the ride fares price based on the route
	estimatedFares := h.service.EstimatePackagesPriceWithRoutes(route)

	// store the ride fares for creating trip later
	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, userID, route)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to generate trip fares: %v", err)
	}

	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}

func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	fareID := req.GetRideFareID()
	userID := req.GetUserID()

	rideFare, err := h.service.GetAndValidateFare(ctx, fareID, userID)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get and validate fare: %v", err)
	}

	trip, err := h.service.CreateTrip(ctx, rideFare)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to create trip: %v", err)
	}

	return &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
	}, nil
}

package main

import (
	"encoding/json"
	"log"
	"net/http"

	grpcclients "github.com/tenteedee/mini-uber/services/api-gateway/grpc_clients"
	"github.com/tenteedee/mini-uber/shared/contracts"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
	var requestBody previewTripRequest

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if requestBody.UserId == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	tripService, err := grpcclients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer tripService.Close()

	tripPreview, err := tripService.Client.PreviewTrip(r.Context(), requestBody.ToProto())
	if err != nil {
		log.Printf("Failed to preview trip: %v", err)
		http.Error(w, "Failed to preview trip: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{
		Data: tripPreview,
	}

	writeJSON(w, http.StatusCreated, response)

}

func handleTripStart(w http.ResponseWriter, r *http.Request) {
	var reqBody startTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	tripService, err := grpcclients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer tripService.Close()

	trip, err := tripService.Client.CreateTrip(r.Context(), reqBody.toProto())
	if err != nil {
		log.Printf("Failed to start a trip: %v", err)
		http.Error(w, "Failed to start trip", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: trip}

	writeJSON(w, http.StatusCreated, response)
}

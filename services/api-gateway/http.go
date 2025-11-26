package main

import (
	"encoding/json"
	"net/http"
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

	// TODO: call trip service

	writeJSON(w, http.StatusCreated, "")

}

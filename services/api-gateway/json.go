package main

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func readJSON(r *http.Request, dest any) error {
	return json.NewDecoder(r.Body).Decode(dest)
}

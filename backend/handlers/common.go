package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/RK-Consulting/skill-sifter/models"
)

// respondWithJSON is a helper function to respond with JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success":false,"message":"Error marshalling JSON"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// respondWithError is a helper function to respond with an error
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, models.ApiResponse{
		Success: false,
		Message: message,
	})
}

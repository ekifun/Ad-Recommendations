package handlers

import (
	"Ad-Recommendations/services"
	"encoding/json"
	"log"
	"net/http"
)

// RecommendationRequest represents the incoming recommendation request data
type RecommendationRequest struct {
	UserID string `json:"user_id"`
}

// RecommendationHandler handles HTTP requests to generate ad recommendations
func RecommendationHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🟢 RecommendationHandler triggered")
	// Parse the JSON body
	var req RecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.UserID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	log.Printf("Processing recommendation request for user: %s", req.UserID)

	// Generate recommendations (this function now fetches user history internally)
	recommendations := services.GenerateRecommendations(req.UserID)

	// Check if recommendations were generated
	if recommendations == nil {
		http.Error(w, "Failed to generate recommendations", http.StatusInternalServerError)
		return
	}

	// Respond with recommendations
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

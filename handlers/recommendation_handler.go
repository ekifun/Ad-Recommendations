package handlers

import (
	"Ad-Recommendations/models"
	"Ad-Recommendations/services"
	"encoding/json"
	"net/http"
)

// RecommendationRequest represents the incoming recommendation request data
type RecommendationRequest struct {
	UserID string `json:"user_id"`
}

// RecommendationHandler handles HTTP requests to generate ad recommendations
func RecommendationHandler(w http.ResponseWriter, r *http.Request) {
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

	// Fetch user history
	playbackHistory, err := services.FetchUserPlaybackHistory(req.UserID)
	if err != nil {
		http.Error(w, "Failed to fetch playback history", http.StatusInternalServerError)
		return
	}

	adClickHistory, err := services.FetchUserAdClickHistory(req.UserID)
	if err != nil {
		http.Error(w, "Failed to fetch ad click history", http.StatusInternalServerError)
		return
	}

	// Fetch ads and embeddings (dummy ads and embeddings used for example)
	ads := []models.Ad{
		{AdID: "1", Category: "Tech", Description: "Latest gadgets and devices"},
		{AdID: "2", Category: "Fitness", Description: "Workout equipment and accessories"},
	}
	embeddings := [][]float64{
		{0.8, 0.6, 0.4},
		{0.7, 0.5, 0.3},
	}

	// Generate recommendations
	recommendations := services.GenerateRecommendations(playbackHistory, adClickHistory, ads, embeddings)

	// Respond with recommendations
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

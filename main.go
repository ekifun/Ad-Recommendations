package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Ad represents a single advertisement.
type Ad struct {
	AdID           string   `json:"ad_id"`
	Category       string   `json:"category"`
	Description    string   `json:"description"`
	TargetAudience []string `json:"target_audience"`
}

// UserProfile represents a user's static preference profile.
type UserProfile struct {
	UserID      string   `json:"user_id"`
	Preferences []string `json:"preferences"`
}

// Static Ad Inventory
var ads = []Ad{
	{"ad1", "Tech", "Ad for a tech gadget", []string{"tech", "gadgets"}},
	{"ad2", "Travel", "Ad for a travel package", []string{"travel", "adventure"}},
	{"ad3", "Food", "Ad for a new restaurant", []string{"food", "restaurants"}},
}

// Static User Profiles
var users = map[string]UserProfile{
	"user1": {"user1", []string{"tech", "gadgets"}},
	"user2": {"user2", []string{"travel", "adventure"}},
	"user3": {"user3", []string{"food", "restaurants"}},
}

// RecommendAds filters ads based on the user's preferences.
func RecommendAds(user UserProfile, ads []Ad) []Ad {
	recommendations := []Ad{}
	for _, ad := range ads {
		if matchesPreferences(user.Preferences, ad.TargetAudience) {
			recommendations = append(recommendations, ad)
		}
	}
	return recommendations
}

// matchesPreferences checks if any user preference matches the ad's target audience.
func matchesPreferences(preferences, targets []string) bool {
	for _, pref := range preferences {
		for _, target := range targets {
			if strings.EqualFold(pref, target) { // Case-insensitive match
				return true
			}
		}
	}
	return false
}

// getRecommendations is the API handler for fetching recommended ads.
func getRecommendations(w http.ResponseWriter, r *http.Request) {
	// Get the userID from query parameters
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// Find the user profile
	user, exists := users[userID]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Generate recommendations
	recommendations := RecommendAds(user, ads)

	// Respond with JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

// main sets up the HTTP server and routes.
func main() {
	http.HandleFunc("/recommend", getRecommendations)

	// Start the HTTP server
	println("Server running on http://localhost:8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		println("Error starting server:", err.Error())
	}
}

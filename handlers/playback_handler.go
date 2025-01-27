package handlers

import (
	"Ad-Recommendations/services"
	"Ad-Recommendations/utils"
	"encoding/json"
	"net/http"
)

// PlaybackHandler handles movie playback logging
func PlaybackHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID   string `json:"user_id"`
		Category string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := services.LogPlayback(input.UserID, input.Category); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to log playback")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Playback logged successfully"})
}

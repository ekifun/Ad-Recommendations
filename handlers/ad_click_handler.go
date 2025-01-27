package handlers

import (
	"Ad-Recommendations/services"
	"Ad-Recommendations/utils"
	"encoding/json"
	"net/http"
)

// AdClickHandler handles ad click events
func AdClickHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID string `json:"userID"`
		AdID   string `json:"adID"`
	}

	// Parse the request body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.LogError("Failed to decode request body: " + err.Error())
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if input.UserID == "" {
		utils.LogError("userID is missing in the request")
		utils.RespondWithError(w, http.StatusBadRequest, "userID is required")
		return
	}
	if input.AdID == "" {
		utils.LogError("adID is missing in the request")
		utils.RespondWithError(w, http.StatusBadRequest, "adID is required")
		return
	}

	// Log the ad click
	if err := services.LogAdClick(input.UserID, input.AdID); err != nil {
		utils.LogError("Failed to log ad click: " + err.Error())
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to log ad click")
		return
	}

	utils.LogInfo("Ad click logged successfully: UserID=" + input.UserID + ", AdID=" + input.AdID)
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Ad click logged successfully"})
}

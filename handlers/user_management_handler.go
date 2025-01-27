package handlers

import (
	"Ad-Recommendations/models"
	"Ad-Recommendations/services"
	"Ad-Recommendations/utils"
	"encoding/json"
	"net/http"
)

// AddUserHandler handles adding a new user profile
func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User

	// Parse the request body
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate userID
	if user.UserID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "userID cannot be empty")
		return
	}

	// Call AddUser service
	if err := services.AddUser(user); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to add user: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "User added successfully"})
}

// UpdateUserHandler handles updating an existing user profile
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := services.UpdateUser(user); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User updated successfully"})
}

// DeleteUserHandler handles deleting a user profile
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing userID")
		return
	}

	if err := services.DeleteUser(userID); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete user: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// GetUserHandler handles retrieving a user profile
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Missing userID")
		return
	}

	user, err := services.GetUser(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve user: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

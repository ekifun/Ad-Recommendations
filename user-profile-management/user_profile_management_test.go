package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test /addUser Endpoint
func TestAddUserEndpoint(t *testing.T) {
	// Mock request payload
	user := UserProfile{
		ID:        "user1",
		Interests: []string{"tech", "travel"},
	}
	userJSON, _ := json.Marshal(user)

	// Create mock HTTP request
	req, err := http.NewRequest("POST", "/addUser", bytes.NewBuffer(userJSON))
	assert.NoError(t, err, "Creating HTTP request should not return an error")

	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(addUserHandler)

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, recorder.Code, "Response code should be 201 Created")
	assert.Contains(t, recorder.Body.String(), "User user1 added successfully", "Response should confirm user addition")

	// Verify user is stored in Redis
	retrievedUser, err := getUserProfileFromRedis("user1")
	assert.NoError(t, err, "Fetching user profile should not return an error")
	assert.Equal(t, user.ID, retrievedUser.ID, "Stored user ID should match")
	assert.ElementsMatch(t, user.Interests, retrievedUser.Interests, "Stored interests should match")
}

// Test /updateUser Endpoint
func TestUpdateUserEndpoint(t *testing.T) {
	// Prepopulate Redis with a user
	preExistingUser := UserProfile{
		ID:        "user1",
		Interests: []string{"tech", "travel"},
	}
	saveUserProfileToRedis(preExistingUser)

	// Mock update payload
	updateInterests := []string{"adventure", "fitness"}
	updateJSON, _ := json.Marshal(updateInterests)

	// Create mock HTTP request
	req, err := http.NewRequest("PUT", "/updateUser?userID=user1", bytes.NewBuffer(updateJSON))
	assert.NoError(t, err, "Creating HTTP request should not return an error")

	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(updateUserHandler)

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Assert response
	assert.Equal(t, http.StatusOK, recorder.Code, "Response code should be 200 OK")
	assert.Contains(t, recorder.Body.String(), "User user1 updated successfully", "Response should confirm user update")

	// Verify user is updated in Redis
	updatedUser, err := getUserProfileFromRedis("user1")
	assert.NoError(t, err, "Fetching updated user profile should not return an error")
	expectedInterests := append(preExistingUser.Interests, updateInterests...)
	assert.ElementsMatch(t, expectedInterests, updatedUser.Interests, "Updated interests should match")
}

// Test /getUser Endpoint
func TestGetUserEndpoint(t *testing.T) {
	// Prepopulate Redis with a user
	mockUser := UserProfile{
		ID:        "user2",
		Interests: []string{"gaming", "sports"},
	}
	saveUserProfileToRedis(mockUser)

	// Create mock HTTP request
	req, err := http.NewRequest("GET", "/getUser?userID=user2", nil)
	assert.NoError(t, err, "Creating HTTP request should not return an error")

	// Create response recorder
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(getUserHandler)

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Assert response
	assert.Equal(t, http.StatusOK, recorder.Code, "Response code should be 200 OK")

	// Parse response body
	var retrievedUser UserProfile
	err = json.Unmarshal(recorder.Body.Bytes(), &retrievedUser)
	assert.NoError(t, err, "Unmarshaling response should not return an error")

	// Verify the retrieved user matches the expected user
	assert.Equal(t, mockUser.ID, retrievedUser.ID, "Retrieved user ID should match")
	assert.ElementsMatch(t, mockUser.Interests, retrievedUser.Interests, "Retrieved interests should match")
}

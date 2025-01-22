package user_profile_management_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test saveUserProfileToRedis and getUserProfileFromRedis
func TestRedisUserProfileFunctions(t *testing.T) {
	redisClient.FlushDB(context.Background()) // Clear Redis for testing

	user := UserProfile{
		ID: "testUser",
		Interests: map[string]int{
			"Tech": 5,
			"Food": 3,
		},
		Age:      25,
		Location: "Test City",
	}

	err := saveUserProfileToRedis(user)
	assert.NoError(t, err)

	retrievedUser, err := getUserProfileFromRedis("testUser")
	assert.NoError(t, err)
	assert.Equal(t, user, retrievedUser)
}

// Test addUserHandler
func TestAddUserHandler(t *testing.T) {
	redisClient.FlushDB(context.Background()) // Clear Redis for testing

	user := UserProfile{
		ID:        "testUser",
		Interests: map[string]int{"Tech": 5},
		Age:       30,
		Location:  "Test City",
	}
	body, _ := json.Marshal(user)

	req := httptest.NewRequest(http.MethodPost, "/addUser", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	AddUserHandler(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "User testUser added successfully")

	retrievedUser, err := getUserProfileFromRedis("testUser")
	assert.NoError(t, err)
	assert.Equal(t, user, retrievedUser)
}

// Test updateUserHandler
func TestUpdateUserHandler(t *testing.T) {
	redisClient.FlushDB(context.Background()) // Clear Redis for testing

	// Initial user profile
	user := UserProfile{
		ID:        "testUser",
		Interests: map[string]int{"Tech": 5},
	}
	saveUserProfileToRedis(user)

	newInterests := map[string]int{"Travel": 3}
	body, _ := json.Marshal(newInterests)

	req := httptest.NewRequest(http.MethodPost, "/updateUser?userID=testUser", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	UpdateUserHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "User testUser updated successfully")

	retrievedUser, err := getUserProfileFromRedis("testUser")
	assert.NoError(t, err)
	assert.Equal(t, 3, retrievedUser.Interests["Travel"])
}

// Test playbackHandler
func TestPlaybackHandler(t *testing.T) {
	redisClient.FlushDB(context.Background()) // Clear Redis for testing

	body := map[string]string{
		"userID":   "testUser",
		"category": "Food",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/playback", bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	PlaybackHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "User profile updated successfully")

	retrievedUser, err := getUserProfileFromRedis("testUser")
	assert.NoError(t, err)
	assert.Equal(t, 1, retrievedUser.Interests["Food"])
}

// Test getUserHandler
func TestGetUserHandler(t *testing.T) {
	redisClient.FlushDB(context.Background()) // Clear Redis for testing

	user := UserProfile{
		ID:        "testUser",
		Interests: map[string]int{"Tech": 5},
	}
	saveUserProfileToRedis(user)

	req := httptest.NewRequest(http.MethodGet, "/getUser?userID=testUser", nil)
	rec := httptest.NewRecorder()

	GetUserHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var retrievedUser UserProfile
	err := json.Unmarshal(rec.Body.Bytes(), &retrievedUser)
	assert.NoError(t, err)
	assert.Equal(t, user, retrievedUser)
}

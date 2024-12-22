package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Mock Data
var mockUser = UserProfile{
	ID:        "user1",
	Interests: []string{"travel", "adventure"},
}

var mockAds = []Ad{
	{
		AdID:           "ad1",
		Category:       "Tech",
		Description:    "Ad for a tech gadget",
		Keywords:       []string{"tech", "gadgets"},
		TargetAudience: []string{"tech", "gadgets"},
		CreatedAt:      time.Now().AddDate(0, 0, -5), // 5 days old
	},
	{
		AdID:           "ad2",
		Category:       "Travel",
		Description:    "Ad for a travel package",
		Keywords:       []string{"travel", "adventure", "vacation"},
		TargetAudience: []string{"travel", "adventure"},
		CreatedAt:      time.Now().AddDate(0, 0, -2), // 2 days old
	},
}

// Test Scoring Logic
func TestScoreAd(t *testing.T) {
	recencyBoost := calculateRecencyBoost(mockAds[1].CreatedAt)
	score := scoreAd(mockAds[1], mockUser.Interests, recencyBoost)

	assert.Equal(t, 4.0, score, "Score should match expected value")
}

// Test Redis Integration
func TestRedisIntegration(t *testing.T) {
	err := saveUserProfileToRedis(mockUser)
	assert.NoError(t, err, "Saving user profile should not return an error")

	user, err := getUserProfileFromRedis("user1")
	assert.NoError(t, err, "Fetching user profile should not return an error")
	assert.Equal(t, mockUser.ID, user.ID, "Fetched user ID should match")
	assert.ElementsMatch(t, mockUser.Interests, user.Interests, "Fetched user interests should match")
}

// Test Database Integration
func TestDatabaseIntegration(t *testing.T) {
	rows, err := dbPool.Query(context.Background(), "SELECT ad_id, category, description, target_audience, keywords, created_at FROM ads")
	assert.NoError(t, err, "Fetching ads should not return an error")
	defer rows.Close()

	ads := []Ad{}
	for rows.Next() {
		var ad Ad
		err := rows.Scan(&ad.AdID, &ad.Category, &ad.Description, &ad.TargetAudience, &ad.Keywords, &ad.CreatedAt)
		assert.NoError(t, err, "Scanning ad should not return an error")
		ads = append(ads, ad)
	}
	assert.Greater(t, len(ads), 0, "Ads table should not be empty")
}

// Test /recommend Endpoint
func TestRecommendEndpoint(t *testing.T) {
	// Mock user profile in Redis
	saveUserProfileToRedis(mockUser)

	// Mock HTTP request
	req, err := http.NewRequest("GET", "/recommend?userID=user1", nil)
	assert.NoError(t, err, "Creating HTTP request should not return an error")

	// Mock HTTP response recorder
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(recommendHandler)
	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Response code should be 200 OK")

	// Validate response body
	var recommendations []Ad
	err = json.Unmarshal(recorder.Body.Bytes(), &recommendations)
	assert.NoError(t, err, "Unmarshaling response should not return an error")
	assert.Greater(t, len(recommendations), 0, "Recommendations should not be empty")
	assert.Equal(t, "ad2", recommendations[0].AdID, "Top recommendation should be 'ad2'")
}

package main

import (
	"Ad-Recommendations/user_profile_management"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

// Mock Redis Client
type MockRedisClient struct{}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	if key == "test_user" {
		mockUserProfile := user_profile_management.UserProfile{
			ID:        "test_user",
			Interests: map[string]int{"Tech": 5, "Travel": 3},
		}
		userJSON, _ := json.Marshal(mockUserProfile)
		return redis.NewStringResult(string(userJSON), nil)
	}
	return redis.NewStringResult("", redis.Nil)
}

// Mock PGX Pool
type MockPGXPool struct{}

func (m *MockPGXPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return &MockRows{
		data: [][]interface{}{
			{"ad1", "Tech", "Ad for a tech gadget", []string{"tech", "gadgets"}, []string{"tech", "gadgets"}, time.Now()},
			{"ad2", "Travel", "Ad for a travel destination", []string{"travel", "vacation"}, []string{"beach", "adventure"}, time.Now()},
		},
		index: -1,
	}, nil
}

// Mock Rows for PGX
type MockRows struct {
	data  [][]interface{}
	index int
}

func (m *MockRows) Next() bool {
	m.index++
	return m.index < len(m.data)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.index < 0 || m.index >= len(m.data) {
		return errors.New("out of range")
	}
	row := m.data[m.index]
	for i := range dest {
		switch v := dest[i].(type) {
		case *string:
			*v = row[i].(string)
		case *[]string:
			*v = row[i].([]string)
		case *time.Time:
			*v = row[i].(time.Time)
		default:
			return errors.New("unsupported type")
		}
	}
	return nil
}

func (m *MockRows) Close() {}

func (m *MockRows) Err() error {
	return nil
}

// Updated CommandTag method to return a dummy pgconn.CommandTag
func (m *MockRows) CommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag("SELECT 2")
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (m *MockRows) Conn() *pgx.Conn {
	return nil
}

func TestRecommendHandler(t *testing.T) {
	// Replace global dependencies with mocks
	redisClient = redis.NewClient(&redis.Options{}) // Dummy Redis client
	redisClient = &MockRedisClient{}
	dbPool = &MockPGXPool{}

	req := httptest.NewRequest("GET", "/recommend?userID=test_user", nil)
	rr := httptest.NewRecorder()

	handler := corsMiddleware(http.HandlerFunc(recommendHandler))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var recommendations []Ad
	err := json.Unmarshal(rr.Body.Bytes(), &recommendations)
	assert.NoError(t, err)
	assert.Len(t, recommendations, 2)
	assert.Equal(t, "Tech", recommendations[0].Category)
	assert.Equal(t, "Travel", recommendations[1].Category)
}

func TestAddUserHandler(t *testing.T) {
	userData := `{"id":"test_user","interests":{"Tech":5},"age":30,"location":"New York"}`
	req := httptest.NewRequest("POST", "/addUser", strings.NewReader(userData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := corsMiddleware(http.HandlerFunc(user_profile_management.AddUserHandler))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Contains(t, rr.Body.String(), "User test_user added successfully")
}

func TestUpdateUserHandler(t *testing.T) {
	updateData := `["Travel", "Food"]`
	req := httptest.NewRequest("PUT", "/updateUser?userID=test_user", strings.NewReader(updateData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := corsMiddleware(http.HandlerFunc(user_profile_management.UpdateUserHandler))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "User test_user updated successfully")
}

func TestDeleteUserHandler(t *testing.T) {
	req := httptest.NewRequest("DELETE", "/deleteUser?userID=test_user", nil)
	rr := httptest.NewRecorder()

	handler := corsMiddleware(http.HandlerFunc(user_profile_management.DeleteUserHandler))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "User test_user deleted successfully")
}

func TestPlaybackHandler(t *testing.T) {
	playbackData := `{"userID":"test_user","category":"Fitness"}`
	req := httptest.NewRequest("POST", "/playback", strings.NewReader(playbackData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := corsMiddleware(http.HandlerFunc(user_profile_management.PlaybackHandler))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "User profile updated successfully")
}

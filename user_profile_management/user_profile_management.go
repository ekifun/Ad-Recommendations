package user_profile_management

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
)

// UserProfile represents the user profile structure
type UserProfile struct {
	ID                 string         `json:"id"`
	Interests          map[string]int `json:"interests"`
	Age                int            `json:"age,omitempty"`
	Location           string         `json:"location,omitempty"`
	RecentlyInteracted []string       `json:"recently_interacted,omitempty"`
}

var redisClient *redis.Client

func init() {
	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
	})
}

// AddUserHandler handles adding a new user profile
func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user UserProfile
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := saveUserProfileToRedis(user); err != nil {
		log.Printf("Error saving user profile: %v", err)
		http.Error(w, "Error saving user profile", http.StatusInternalServerError)
		return
	}

	log.Printf("User profile added successfully: %s", user.ID)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User added successfully")
}

// UpdateUserHandler updates an existing user profile
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	var payload struct {
		Interests []string `json:"interests"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Fetch the user profile
	user, err := GetUserProfileFromRedis(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update interests
	for _, interest := range payload.Interests {
		user.Interests[interest]++
	}

	// Save updated profile
	if err := saveUserProfileToRedis(user); err != nil {
		log.Println("Error saving user profile:", err)
		http.Error(w, "Error saving user profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User %s updated successfully", userID)
}

// DeleteUserHandler deletes a user profile
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	if err := redisClient.Del(context.Background(), userID).Err(); err != nil {
		log.Printf("Error deleting user profile: %v", err)
		http.Error(w, "Error deleting user profile", http.StatusInternalServerError)
		return
	}

	log.Printf("User profile deleted successfully: %s", userID)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "User deleted successfully")
}

func CalculateRecencyBoost(category string, recentlyInteracted []string) int {
	for i, c := range recentlyInteracted {
		if c == category {
			return 10 - i // Higher score for more recent interactions
		}
	}
	return 0
}

// PlaybackHandler updates user interests based on playback category
func PlaybackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		UserID   string `json:"userID"`
		Category string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := GetUserProfileFromRedis(request.UserID)
	if err != nil {
		if err.Error() == "user profile not found" {
			user = UserProfile{
				ID:                 request.UserID,
				Interests:          make(map[string]int),
				RecentlyInteracted: []string{},
			}
		} else {
			http.Error(w, "Error fetching user profile", http.StatusInternalServerError)
			return
		}
	}

	user.Interests[request.Category]++

	// Update RecentlyInteracted
	user.RecentlyInteracted = append([]string{request.Category}, user.RecentlyInteracted...)
	if len(user.RecentlyInteracted) > 5 {
		user.RecentlyInteracted = user.RecentlyInteracted[:5]
	}

	if err := saveUserProfileToRedis(user); err != nil {
		http.Error(w, "Error saving user profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "User profile updated successfully")
}

// Helper functions for user profile management
func saveUserProfileToRedis(user UserProfile) error {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user profile: %v", err)
	}

	err = redisClient.Set(context.Background(), user.ID, userJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save user profile to Redis: %v", err)
	}
	return nil
}

func GetUserProfileFromRedis(userID string) (UserProfile, error) {
	userJSON, err := redisClient.Get(context.Background(), userID).Result()
	if err != nil {
		if err == redis.Nil {
			return UserProfile{}, fmt.Errorf("user profile not found")
		}
		return UserProfile{}, fmt.Errorf("failed to fetch user profile: %v", err)
	}

	var user UserProfile
	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		return UserProfile{}, fmt.Errorf("failed to unmarshal user profile: %v", err)
	}

	return user, nil
}

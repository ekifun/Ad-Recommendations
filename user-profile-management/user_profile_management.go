package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
)

type UserProfile struct {
	ID             string   `json:"id"`
	Interests      []string `json:"interests"`
	Age            int      `json:"age,omitempty"`
	Location       string   `json:"location,omitempty"`
	RecentlyViewed []string `json:"recently_viewed,omitempty"`
}

var redisClient *redis.Client

func init() {
	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
	})
}

func saveUserProfileToRedis(user UserProfile) error {
	// Convert the user profile to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user profile: %v", err)
	}

	// Store the JSON in Redis with the user ID as the key
	err = redisClient.Set(context.Background(), user.ID, userJSON, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save user profile to Redis: %v", err)
	}
	return nil
}

func getUserProfileFromRedis(userID string) (UserProfile, error) {
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

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	var user UserProfile
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Save user profile to Redis
	err := saveUserProfileToRedis(user)
	if err != nil {
		log.Println("Error saving user profile:", err)
		http.Error(w, "Error saving user profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "User %s added successfully", user.ID)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract userID from query parameters
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// Fetch the existing user profile from Redis
	user, err := getUserProfileFromRedis(userID)
	if err != nil {
		log.Println("Error fetching user profile:", err)
		if err.Error() == "user profile not found" {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error fetching user profile", http.StatusInternalServerError)
		}
		return
	}

	// Decode the new interests from the request body
	var newInterests []string
	if err := json.NewDecoder(r.Body).Decode(&newInterests); err != nil {
		log.Println("Error decoding interests:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Append new interests to the existing ones, avoiding duplicates
	existingInterests := make(map[string]bool)
	for _, interest := range user.Interests {
		existingInterests[interest] = true
	}
	for _, interest := range newInterests {
		if !existingInterests[interest] {
			user.Interests = append(user.Interests, interest)
		}
	}

	// Save the updated profile back to Redis
	if err := saveUserProfileToRedis(user); err != nil {
		log.Println("Error saving updated user profile:", err)
		http.Error(w, "Error saving user profile", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User %s updated successfully", user.ID)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// Fetch the user profile from Redis
	userJSON, err := redisClient.Get(context.Background(), userID).Result()
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(userJSON))
}

func main() {
	// Other routes
	http.HandleFunc("/addUser", addUserHandler)
	http.HandleFunc("/updateUser", updateUserHandler) // Ensure this is added
	http.HandleFunc("/getUser", getUserHandler)

	// Start server
	port := ":8082"
	log.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

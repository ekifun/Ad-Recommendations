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

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	var user UserProfile
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Save user profile to Redis
	userJSON, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Error encoding user data", http.StatusInternalServerError)
		return
	}
	redisClient.Set(context.Background(), user.ID, userJSON, 0)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "User %s added successfully", user.ID)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// Fetch the existing user profile
	userJSON, err := redisClient.Get(context.Background(), userID).Result()
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var user UserProfile
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		http.Error(w, "Error decoding user data", http.StatusInternalServerError)
		return
	}

	// Update interests
	var newInterests []string
	if err := json.NewDecoder(r.Body).Decode(&newInterests); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	user.Interests = append(user.Interests, newInterests...)

	// Save updated user profile
	updatedUserJSON, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Error encoding updated user data", http.StatusInternalServerError)
		return
	}
	redisClient.Set(context.Background(), user.ID, updatedUserJSON, 0)

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

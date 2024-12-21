package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
)

type UserProfile struct {
	ID        string   `json:"id"`
	Interests []string `json:"interests"`
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
	userJSON, _ := json.Marshal(user)
	redisClient.Set(context.Background(), user.ID, userJSON, 0)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "User %s added successfully", user.ID)
}

func main() {
	http.HandleFunc("/addUser", addUserHandler)

	// Start the server
	port := ":8082"
	fmt.Printf("Server running on http://localhost%s\n", port)
	http.ListenAndServe(port, nil)
}

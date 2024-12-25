package main

import (
	"Ad-Recommendations/user_profile_management" // Importing the user profile management package
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Ad struct {
	AdID           string    `json:"ad_id"`
	Category       string    `json:"category"`
	Description    string    `json:"description"`
	TargetAudience []string  `json:"target_audience"`
	Keywords       []string  `json:"keywords"`
	CreatedAt      time.Time `json:"created_at"`
}

var (
	dbPool      *pgxpool.Pool
	redisClient *redis.Client
)

func init() {
	var err error

	// Initialize PostgreSQL connection
	dbPool, err = pgxpool.Connect(context.Background(), "postgresql://demo_user:demo_password@localhost:5432/ad_inventory")
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Initialize Redis connection
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func recommendHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Recommendation endpoint hit")

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// Fetch user profile
	log.Printf("Fetching user profile for userID: %s", userID)
	user, err := user_profile_management.GetUserProfileFromRedis(userID)
	if err != nil {
		log.Printf("Error fetching user profile for userID %s: %v", userID, err)
		http.Error(w, "User profile not found", http.StatusNotFound)
		return
	}
	log.Printf("Fetched User Profile: %+v", user)

	// Fetch ads from database
	rows, err := dbPool.Query(context.Background(), "SELECT ad_id, category, description, target_audience, keywords, created_at FROM ads")
	if err != nil {
		log.Printf("Error executing query: %v", err)
		http.Error(w, "Failed to fetch ads", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	recommendations := []Ad{}
	for rows.Next() {
		var ad Ad
		if err := rows.Scan(&ad.AdID, &ad.Category, &ad.Description, &ad.TargetAudience, &ad.Keywords, &ad.CreatedAt); err != nil {
			log.Printf("Error scanning ad row: %v", err)
			continue
		}
		recommendations = append(recommendations, ad)
	}

	if len(recommendations) == 0 {
		log.Println("No ads found")
		http.Error(w, "No ads available", http.StatusNotFound)
		return
	}

	log.Printf("Returning %d ads", len(recommendations))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// corsMiddleware wraps an HTTP handler to include CORS support
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method == http.MethodOptions {
			// Preflight request: respond with OK status
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	http.Handle("/recommend", corsMiddleware(http.HandlerFunc(recommendHandler)))
	http.Handle("/addUser", corsMiddleware(http.HandlerFunc(user_profile_management.AddUserHandler)))
	http.Handle("/updateUser", corsMiddleware(http.HandlerFunc(user_profile_management.UpdateUserHandler)))
	http.Handle("/deleteUser", corsMiddleware(http.HandlerFunc(user_profile_management.DeleteUserHandler)))
	http.Handle("/playback", corsMiddleware(http.HandlerFunc(user_profile_management.PlaybackHandler)))

	port := ":8082"
	log.Printf("Server running on http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Structs
type UserProfile struct {
	ID        string   `json:"id"`
	Interests []string `json:"interests"`
	Age       int      `json:"age,omitempty"`
	Location  string   `json:"location,omitempty"`
}

type Ad struct {
	AdID           string   `json:"ad_id"`
	Category       string   `json:"category"`
	Description    string   `json:"description"`
	TargetAudience []string `json:"target_audience"`
	Keywords       []string `json:"keywords"`
	CreatedAt      time.Time
	Score          float64 `json:"score,omitempty"`
}

// Global Variables
var dbPool *pgxpool.Pool
var redisClient *redis.Client

func init() {
	// Initialize PostgreSQL connection
	var err error
	dbPool, err = pgxpool.Connect(context.Background(), "postgresql://john:my_db_password@localhost:5432/ad_inventory")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Initialize Redis connection
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

// Handlers

// Recommend Handler
func recommendHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// Fetch user profile from Redis
	userJSON, err := redisClient.Get(context.Background(), userID).Result()
	if err != nil {
		log.Println("Error fetching ads:", err)
		http.Error(w, "Error fetching ads", http.StatusInternalServerError)
		return
	}

	var user UserProfile
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		log.Println("Error decoding user profile:", err)
		http.Error(w, "Error decoding user profile", http.StatusInternalServerError)
		return
	}

	// Fetch all ads from PostgreSQL
	rows, err := dbPool.Query(context.Background(), "SELECT ad_id, category, description, target_audience, keywords, created_at FROM ads")
	if err != nil {
		log.Println("Error fetching ads:", err)
		http.Error(w, "Error fetching ads", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Score each ad
	recommendations := []Ad{}
	for rows.Next() {
		var ad Ad
		err := rows.Scan(&ad.AdID, &ad.Category, &ad.Description, &ad.TargetAudience, &ad.Keywords, &ad.CreatedAt)
		if err != nil {
			log.Println("Error scanning ad row:", err)
			continue
		}

		// Calculate recency boost
		recencyBoost := calculateRecencyBoost(ad.CreatedAt)

		// Calculate ad score
		ad.Score = scoreAd(ad, user.Interests, recencyBoost)
		if ad.Score > 0 {
			recommendations = append(recommendations, ad)
		}
	}

	// Sort ads by score in descending order
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	// Return top recommendations
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

// Utility Functions

// Scoring function for an ad
func scoreAd(ad Ad, userInterests []string, recencyBoost float64) float64 {
	keywordMatchScore := 0.0
	for _, interest := range userInterests {
		for _, keyword := range ad.Keywords {
			if interest == keyword {
				keywordMatchScore += 1.0 // Increment score for each matching keyword
			}
		}
	}

	// Apply category-specific weights (example: boost Travel ads)
	categoryWeight := 1.0
	if ad.Category == "Travel" {
		categoryWeight = 1.5
	}

	// Final score calculation
	finalScore := (keywordMatchScore * categoryWeight) + recencyBoost
	return finalScore
}

// Recency boost calculation
func calculateRecencyBoost(createdAt time.Time) float64 {
	daysOld := time.Since(createdAt).Hours() / 24
	if daysOld < 7 {
		return 1.0 // Full boost for ads less than a week old
	} else if daysOld < 30 {
		return 0.5 // Reduced boost for ads less than a month old
	}
	return 0.0 // No boost for older ads
}

func main() {
	// Register handlers
	http.HandleFunc("/recommend", recommendHandler)

	// Start server
	port := ":8082"
	log.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

package main

import (
	"Ad-Recommendations/user_profile_management"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Ad struct {
	AdID           string    `json:"ad_id"`
	Category       string    `json:"category"`
	Description    string    `json:"description"`
	TargetAudience []string  `json:"target_audience"`
	Keywords       []string  `json:"keywords"`
	Embedding      []float64 `json:"embedding,omitempty"` // Add for embedding storage
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
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// Fetch user profile
	user, err := user_profile_management.GetUserProfileFromRedis(userID)
	if err != nil {
		http.Error(w, "User profile not found", http.StatusNotFound)
		return
	}

	// Fetch ads from database
	rows, err := dbPool.Query(context.Background(), "SELECT ad_id, category, description, target_audience, keywords FROM ads")
	if err != nil {
		http.Error(w, "Failed to fetch ads", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	ads := []Ad{}
	texts := []string{}
	for rows.Next() {
		var ad Ad
		if err := rows.Scan(&ad.AdID, &ad.Category, &ad.Description, &ad.TargetAudience, &ad.Keywords); err == nil {
			ads = append(ads, ad)
			texts = append(texts, ad.Description+" "+strings.Join(ad.Keywords, " "))
		}
	}

	// Generate embeddings using Python service
	embeddings, err := generateEmbeddings(texts, "http://localhost:5001/generate-bert")
	if err != nil {
		http.Error(w, "Failed to generate embeddings", http.StatusInternalServerError)
		return
	}

	// Compute cosine similarity
	userVector := generateUserProfileVector(user, embeddings, ads)
	recommendations := rankRecommendations(ads, embeddings, userVector)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

func generateEmbeddings(texts []string, url string) ([][]float64, error) {
	data := map[string]interface{}{"texts": texts}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var embeddings [][]float64
	err = json.NewDecoder(resp.Body).Decode(&embeddings)
	return embeddings, err
}

func generateUserProfileVector(user user_profile_management.UserProfile, embeddings [][]float64, ads []Ad) []float64 {
	userVector := make([]float64, len(embeddings[0])) // Initialize a vector with the same dimension as embeddings.

	for _, category := range user.RecentlyInteracted {
		for i, ad := range ads {
			if category == ad.Category { // Match category with the ad's category.
				for j := range userVector {
					userVector[j] += embeddings[i][j] // Accumulate the embedding of the matching ad.
				}
			}
		}
	}
	return userVector
}

func rankRecommendations(ads []Ad, embeddings [][]float64, userVector []float64) []Ad {
	scores := []struct {
		Ad    Ad
		Score float64
	}{}

	for i, adEmbedding := range embeddings {
		score := cosineSimilarity(userVector, adEmbedding)
		scores = append(scores, struct {
			Ad    Ad
			Score float64
		}{ads[i], score})
	}

	// Sort by score in descending order
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Filter out duplicates using a map
	seen := make(map[string]bool)
	recommendations := []Ad{}
	for _, entry := range scores {
		if !seen[entry.Ad.AdID] {
			recommendations = append(recommendations, entry.Ad)
			seen[entry.Ad.AdID] = true
		}
		if len(recommendations) >= 5 { // Limit to top 5
			break
		}
	}

	return recommendations
}

func cosineSimilarity(vecA, vecB []float64) float64 {
	var dotProduct, normA, normB float64
	for i := range vecA {
		dotProduct += vecA[i] * vecB[i]
		normA += vecA[i] * vecA[i]
		normB += vecB[i] * vecB[i]
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method == http.MethodOptions {
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

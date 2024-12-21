package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
)

var dbPool *pgxpool.Pool

type Ad struct {
	AdID           string   `json:"ad_id"`
	Category       string   `json:"category"`
	Description    string   `json:"description"`
	TargetAudience []string `json:"target_audience"`
}

func main() {
	// PostgreSQL connection setup
	var err error
	dbPool, err = pgxpool.Connect(context.Background(), "postgresql://john:my_db_password@localhost:5432/ad_inventory")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	// Routes
	http.HandleFunc("/addAd", addAdHandler)
	http.HandleFunc("/deleteAd", deleteAdHandler)
	http.HandleFunc("/updateAd", updateAdHandler)

	// Start server
	port := ":8082"
	log.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// Handlers

// Add Ad
func addAdHandler(w http.ResponseWriter, r *http.Request) {
	var ad Ad
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		log.Println("Failed to decode request body:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO ads (ad_id, category, description, target_audience) VALUES ($1, $2, $3, $4)`
	_, dbErr := dbPool.Exec(context.Background(), query, ad.AdID, ad.Category, ad.Description, ad.TargetAudience)
	if dbErr != nil {
		log.Println("Database error:", dbErr)
		http.Error(w, "Error adding ad", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Ad added successfully")
}

// Delete Ad
func deleteAdHandler(w http.ResponseWriter, r *http.Request) {
	adID := r.URL.Query().Get("adID")
	if adID == "" {
		http.Error(w, "Missing adID parameter", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM ads WHERE ad_id = $1`
	_, err := dbPool.Exec(context.Background(), query, adID)
	if err != nil {
		http.Error(w, "Error deleting ad", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ad deleted successfully")
}

// Update Ad
func updateAdHandler(w http.ResponseWriter, r *http.Request) {
	var ad Ad
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	query := `UPDATE ads SET category = $1, description = $2, target_audience = $3 WHERE ad_id = $4`
	_, err := dbPool.Exec(context.Background(), query, ad.Category, ad.Description, ad.TargetAudience, ad.AdID)
	if err != nil {
		http.Error(w, "Error updating ad", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ad updated successfully")
}

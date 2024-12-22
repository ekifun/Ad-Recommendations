package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/assert"
)

func setupDatabase() {
	var err error
	dbPool, err = pgxpool.Connect(context.Background(), "postgresql://john:my_db_password@localhost:5432/ad_inventory_test")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
}

func setupTestTable() {
	// Create the ads table
	_, err := dbPool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS ads (
			ad_id VARCHAR(50) PRIMARY KEY,
			category VARCHAR(50),
			description TEXT,
			target_audience TEXT[],
			keywords TEXT[]
		)
	`)
	if err != nil {
		log.Fatalf("Error creating ads table: %v\n", err)
	}

	// Insert mock data
	_, err = dbPool.Exec(context.Background(), `
		INSERT INTO ads (ad_id, category, description, target_audience, keywords)
		VALUES 
		('ad1', 'Tech', 'Ad for a tech gadget', ARRAY['tech', 'gadgets'], ARRAY['tech', 'innovation']),
		('ad2', 'Travel', 'Ad for a travel package', ARRAY['travel', 'adventure'], ARRAY['vacation', 'adventure']),
		('ad3', 'Fitness', 'Ad for a fitness program', ARRAY['fitness', 'health'], ARRAY['wellness', 'exercise'])
		ON CONFLICT (ad_id) DO NOTHING
	`)
	if err != nil {
		log.Fatalf("Error inserting mock data: %v\n", err)
	}
}

func clearAdsTable() {
	_, err := dbPool.Exec(context.Background(), "TRUNCATE ads RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("Error clearing ads table: %v\n", err)
	}
}

// Test /addAd Endpoint
func TestAddAdEndpoint(t *testing.T) {
	// Clear the table before running the test
	clearAdsTable()

	// Mock request payload
	ad := Ad{
		AdID:           "ad1",
		Category:       "Tech",
		Description:    "Ad for a tech gadget",
		TargetAudience: []string{"tech", "gadgets"},
		Keywords:       []string{"tech", "innovation"},
	}
	adJSON, _ := json.Marshal(ad)

	// Create mock HTTP request
	req, err := http.NewRequest("POST", "/addAd", bytes.NewBuffer(adJSON))
	assert.NoError(t, err, "Creating HTTP request should not return an error")
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(addAdHandler)

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, recorder.Code, "Response code should be 201 Created")
	assert.Contains(t, recorder.Body.String(), "Ad added successfully", "Response should confirm ad addition")

	// Verify ad is stored in PostgreSQL
	row := dbPool.QueryRow(context.Background(), "SELECT ad_id, category, description FROM ads WHERE ad_id = $1", "ad1")
	var storedAd Ad
	err = row.Scan(&storedAd.AdID, &storedAd.Category, &storedAd.Description)
	assert.NoError(t, err, "Fetching ad from database should not return an error")
	assert.Equal(t, ad.AdID, storedAd.AdID, "Stored ad ID should match")
	assert.Equal(t, ad.Category, storedAd.Category, "Stored category should match")
	assert.Equal(t, ad.Description, storedAd.Description, "Stored description should match")
}

// Test /deleteAd Endpoint
func TestDeleteAdEndpoint(t *testing.T) {
	// Prepopulate PostgreSQL with a mock ad
	_, err := dbPool.Exec(context.Background(),
		"INSERT INTO ads (ad_id, category, description, keywords) VALUES ($1, $2, $3, $4) ON CONFLICT (ad_id) DO NOTHING",
		"ad2", "Travel", "Ad for a travel package", []string{"travel", "vacation"})
	assert.NoError(t, err, "Inserting ad into database should not return an error")

	// Create mock HTTP request
	req, err := http.NewRequest("DELETE", "/deleteAd?adID=ad2", nil)
	assert.NoError(t, err, "Creating HTTP request should not return an error")

	// Create response recorder
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(deleteAdHandler)

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Assert response
	assert.Equal(t, http.StatusOK, recorder.Code, "Response code should be 200 OK")
	assert.Contains(t, recorder.Body.String(), "Ad deleted successfully", "Response should confirm ad deletion")

	// Verify ad is deleted from PostgreSQL
	row := dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM ads WHERE ad_id = $1", "ad2")
	var count int
	err = row.Scan(&count)
	assert.NoError(t, err, "Fetching ad count from database should not return an error")
	assert.Equal(t, 0, count, "Ad should be deleted")
}

// Test /updateAd Endpoint
func TestUpdateAdEndpoint(t *testing.T) {
	// Prepopulate PostgreSQL with a mock ad
	_, err := dbPool.Exec(context.Background(),
		"INSERT INTO ads (ad_id, category, description, keywords) VALUES ($1, $2, $3, $4) ON CONFLICT (ad_id) DO NOTHING",
		"ad3", "Tech", "Old description", []string{"tech"})
	assert.NoError(t, err, "Inserting ad into database should not return an error")

	// Mock update payload
	updateAd := Ad{
		AdID:        "ad3",
		Category:    "Tech",
		Description: "Updated description",
		Keywords:    []string{"tech", "gadgets"},
	}
	updateAdJSON, _ := json.Marshal(updateAd)

	// Create mock HTTP request
	req, err := http.NewRequest("PUT", "/updateAd", bytes.NewBuffer(updateAdJSON))
	assert.NoError(t, err, "Creating HTTP request should not return an error")
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(updateAdHandler)

	// Call the handler
	handler.ServeHTTP(recorder, req)

	// Assert response
	assert.Equal(t, http.StatusOK, recorder.Code, "Response code should be 200 OK")
	assert.Contains(t, recorder.Body.String(), "Ad updated successfully", "Response should confirm ad update")

	// Verify ad is updated in PostgreSQL
	row := dbPool.QueryRow(context.Background(), "SELECT description, keywords FROM ads WHERE ad_id = $1", "ad3")
	var updatedDescription string
	var updatedKeywords []string
	err = row.Scan(&updatedDescription, &updatedKeywords)
	assert.NoError(t, err, "Fetching updated ad from database should not return an error")
	assert.Equal(t, "Updated description", updatedDescription, "Updated description should match")
	assert.ElementsMatch(t, []string{"tech", "gadgets"}, updatedKeywords, "Updated keywords should match")
}

func TestMain(m *testing.M) {
	setupDatabase()
	setupTestTable()
	defer dbPool.Close()
	m.Run()
}

package main

import (
	"Ad-Recommendations/db"
	"Ad-Recommendations/handlers"
	"Ad-Recommendations/utils"
	"log"
	"net/http"
	"os"
)

func main() {
	// Log server start
	utils.LogInfo("Initializing Ad Recommendation System")

	// Load environment variables (e.g., AWS credentials)
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion := os.Getenv("AWS_REGION")

	if awsAccessKey == "" || awsSecretKey == "" || awsRegion == "" {
		utils.LogError("AWS credentials or region not set. Please configure them using environment variables.")
		log.Fatal("Failed to start application: missing AWS configuration.")
	}

	// Initialize the DynamoDB client
	db.InitDynamoDB()

	// Ensure required DynamoDB tables are available
	if err := db.EnsureTables(); err != nil {
		utils.LogError("Failed to ensure DynamoDB tables: " + err.Error())
		log.Fatal("Failed to start application: unable to initialize DynamoDB tables.")
	}

	// Setup HTTP handlers
	http.Handle("/recommend", utils.CorsMiddleware(http.HandlerFunc(handlers.RecommendationHandler)))
	http.Handle("/playback", utils.CorsMiddleware(http.HandlerFunc(handlers.PlaybackHandler)))
	http.Handle("/ad-click", utils.CorsMiddleware(http.HandlerFunc(handlers.AdClickHandler)))

	// User management (optional, for dynamic user management)
	http.Handle("/add-user", utils.CorsMiddleware(http.HandlerFunc(handlers.AddUserHandler)))
	http.Handle("/update-user", utils.CorsMiddleware(http.HandlerFunc(handlers.UpdateUserHandler)))
	http.Handle("/delete-user", utils.CorsMiddleware(http.HandlerFunc(handlers.DeleteUserHandler)))

	// Start the server
	port := ":8082"
	utils.LogInfo("Server running on http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

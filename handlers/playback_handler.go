package handlers

import (
	"Ad-Recommendations/services"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type PlaybackEvent struct {
	UserID        string `json:"user_id"`
	MovieCategory string `json:"movie_category"`
	Timestamp     string `json:"timestamp"`
}

func PlaybackHandler(w http.ResponseWriter, r *http.Request) {
	var playbackEvent PlaybackEvent

	err := json.NewDecoder(r.Body).Decode(&playbackEvent)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Println("❌ Failed to decode request body:", err)
		return
	}

	if playbackEvent.UserID == "" || playbackEvent.MovieCategory == "" {
		http.Error(w, "Missing user_id or movie_category", http.StatusBadRequest)
		log.Println("❌ Missing required fields: user_id or movie_category")
		return
	}

	if playbackEvent.Timestamp == "" {
		playbackEvent.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	log.Println("🟢 Logging playback to DynamoDB:", playbackEvent)

	err = services.LogPlayback(playbackEvent.UserID, playbackEvent.MovieCategory)
	if err != nil {
		http.Error(w, "Failed to record playback event", http.StatusInternalServerError)
		log.Println("❌ Error saving playback:", err)
		return
	}

	log.Println("✅ Playback recorded successfully:", playbackEvent)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Playback recorded successfully"}`))
}

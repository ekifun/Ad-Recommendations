package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

// Flask AI Server URL
const AI_SERVICE_URL = "http://localhost:5001"

// Handle API Calls to AI Service
func callAIService(endpoint string, texts []string) ([][]float64, error) {
	payload, err := json.Marshal(map[string]interface{}{"texts": texts})
	if err != nil {
		log.Printf("❌ Error: Failed to marshal request for %s: %v", endpoint, err)
		return nil, err
	}

	resp, err := http.Post(AI_SERVICE_URL+endpoint, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("🚨 AI Service Unavailable (%s): %v", endpoint, err)
		return nil, err
	}
	defer resp.Body.Close()

	// Handle HTTP errors (non-200 responses)
	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ AI Service Error (%s): HTTP %d", endpoint, resp.StatusCode)
		return nil, errors.New("AI service returned an error")
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Error: Failed to read response from %s: %v", endpoint, err)
		return nil, err
	}

	// Parse embeddings
	var embeddings [][]float64
	if err := json.Unmarshal(body, &embeddings); err != nil {
		log.Printf("❌ Error: Failed to unmarshal response from %s: %v", endpoint, err)
		return nil, err
	}

	log.Printf("✅ Successfully generated embeddings via %s", endpoint)
	return embeddings, nil
}

// Generate TF-IDF Embeddings
func GenerateTFIDFEmbeddings(texts []string) ([][]float64, error) {
	return callAIService("/generate-tfidf", texts)
}

// Generate BERT Embeddings
func GenerateBERTEmbeddings(texts []string) ([][]float64, error) {
	return callAIService("/generate-bert", texts)
}

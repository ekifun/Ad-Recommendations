package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

// GenerateEmbeddings generates embeddings for the given texts using an external embedding service
func GenerateEmbeddings(texts []string, serviceURL string) ([][]float64, error) {
	// Prepare the request payload
	payload := map[string]interface{}{
		"texts": texts,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v", err)
		return nil, err
	}

	// Send the request to the embedding service
	resp, err := http.Post(serviceURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send request to embedding service: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Embedding service returned status %d: %s", resp.StatusCode, body)
		return nil, errors.New("failed to generate embeddings")
	}

	var embeddings [][]float64
	err = json.Unmarshal(body, &embeddings)
	if err != nil {
		log.Printf("Failed to unmarshal embeddings: %v", err)
		return nil, err
	}

	return embeddings, nil
}

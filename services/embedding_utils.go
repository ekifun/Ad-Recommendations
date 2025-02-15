package services

import (
	"Ad-Recommendations/models" // Import the models package to access the Ad type
	"log"
	"math"
)

// CosineSimilarity calculates the cosine similarity between two vectors
func CosineSimilarity(vecA, vecB []float64) float64 {
	var dotProduct, normA, normB float64
	for i := range vecA {
		dotProduct += vecA[i] * vecB[i]
		normA += vecA[i] * vecA[i]
		normB += vecB[i] * vecB[i]
	}
	if normA == 0 || normB == 0 {
		return 0 // Avoid division by zero
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// GenerateUserVector generates a user vector based on interaction history and ad embeddings
func GenerateUserVector(playbackHistory []string, adClickHistory []string, ads []models.Ad, embeddings [][]float64) []float64 {
	userVector := make([]float64, len(embeddings[0]))

	// Ensure the embeddings list is correctly indexed
	if len(ads) > len(embeddings) {
		log.Printf("❌ Error: Ads (%d) exceed available embeddings (%d). Adjusting...", len(ads), len(embeddings))
		embeddings = padEmbeddings(embeddings, len(ads))
	}

	for i, ad := range ads {
		if i >= len(embeddings) {
			log.Printf("⚠️ Skipping ad %s due to missing embedding", ad.AdID)
			continue
		}

		for j := range userVector {
			userVector[j] += embeddings[i][j] // Avoids index out of range
		}
	}
	return normalizeVector(userVector)
}

// normalizeVector normalizes a vector to unit length
func normalizeVector(vec []float64) []float64 {
	norm := 0.0
	for _, v := range vec {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return vec // Return original vector if norm is 0
	}

	for i := range vec {
		vec[i] /= norm
	}

	return vec
}

// ComputeUserVector averages all embeddings from user playback history
func ComputeUserVector(historyEmbeddings [][]float64) []float64 {
	const embeddingSize = 768 // Adjust based on actual embedding vector size

	if len(historyEmbeddings) == 0 {
		log.Println("⚠️ No embeddings found for user history. Using default vector.")
		return make([]float64, embeddingSize)
	}

	// Compute mean embedding
	userVector := make([]float64, embeddingSize)
	for _, embedding := range historyEmbeddings {
		for i, value := range embedding {
			userVector[i] += value
		}
		log.Println("Generated Embedding:", embedding)
	}

	// Normalize by dividing by count
	for i := range userVector {
		userVector[i] /= float64(len(historyEmbeddings))
	}

	log.Printf("✅ Computed user embedding vector successfully: %v", userVector)
	return userVector
}

// padEmbeddings ensures that all embeddings have a fixed size
func padEmbeddings(embeddings [][]float64, requiredSize int) [][]float64 {
	currentSize := len(embeddings)
	embeddingLength := len(embeddings[0]) // Assuming all embeddings have the same length

	for i := currentSize; i < requiredSize; i++ {
		zeroVector := make([]float64, embeddingLength)
		embeddings = append(embeddings, zeroVector)
	}

	log.Printf("✅ Adjusted embeddings to match ads: %d embeddings for %d ads", len(embeddings), requiredSize)
	return embeddings
}

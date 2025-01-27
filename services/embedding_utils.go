package services

import (
	"Ad-Recommendations/models" // Import the models package to access the Ad type
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
func GenerateUserVector(playbackHistory, adClickHistory []string, ads []models.Ad, embeddings [][]float64) []float64 {
	userVector := make([]float64, len(embeddings[0]))

	// Accumulate embeddings for categories in playback history
	for _, category := range playbackHistory {
		for i, ad := range ads {
			if ad.Category == category {
				for j := range userVector {
					userVector[j] += embeddings[i][j]
				}
			}
		}
	}

	// Additional weighting for ad clicks
	for _, adID := range adClickHistory {
		for i, ad := range ads {
			if ad.AdID == adID {
				for j := range userVector {
					userVector[j] += embeddings[i][j] * 1.5 // Apply weight for clicks
				}
			}
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

package services

import (
	"Ad-Recommendations/db"
	"Ad-Recommendations/models"
	"context"
	"log"
	"math"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// FetchMappedAdCategories retrieves ad category mappings from DynamoDB
func FetchMappedAdCategories(movieCategory string) (map[string]float64, error) {
	log.Printf("🔍 Fetching mapped ad categories for movie category: %s", movieCategory)

	output, err := db.DynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(db.CategoryMappingTableName),
		Key: map[string]types.AttributeValue{
			"movie_category": &types.AttributeValueMemberS{Value: movieCategory},
		},
	})

	if err != nil {
		log.Printf("❌ DynamoDB error fetching category mapping for %s: %v", movieCategory, err)
		return nil, err
	}

	if output.Item == nil || len(output.Item) == 0 {
		log.Printf("⚠️ No category mapping found for movie category: %s", movieCategory)
		return map[string]float64{}, nil
	}

	mappedCategories := make(map[string]float64)

	// Extract mapped categories
	if categoriesAttr, ok := output.Item["ad_categories"].(*types.AttributeValueMemberSS); ok {
		for _, category := range categoriesAttr.Value {
			mappedCategories[category] = 1.0 // Default weight
		}
	}

	// Extract weight if available
	if weightAttr, ok := output.Item["weight"].(*types.AttributeValueMemberN); ok {
		weightValue, err := strconv.ParseFloat(weightAttr.Value, 64)
		if err == nil {
			for key := range mappedCategories {
				mappedCategories[key] = weightValue
			}
		}
	}

	log.Printf("✅ Mapped categories with weights for %s: %v", movieCategory, mappedCategories)
	return mappedCategories, nil
}

// FetchUserPlaybackHistory retrieves the playback history for a user
func FetchUserPlaybackHistory(userID string) ([]string, error) {
	log.Printf("🔍 Fetching playback history for user: %s", userID)

	output, err := db.DynamoClient.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(db.PlaybackTableName),
		KeyConditionExpression: aws.String("user_id = :userID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID},
		},
	})

	if err != nil {
		log.Printf("❌ Failed to query playback history: %v", err)
		return nil, err
	}

	playbackHistory := []string{}
	for _, item := range output.Items {
		if category, ok := item["category"].(*types.AttributeValueMemberS); ok {
			playbackHistory = append(playbackHistory, category.Value)
		}
	}

	log.Printf("✅ Retrieved playback history for %s: %v", userID, playbackHistory)
	return playbackHistory, nil
}

// Normalize category scores and rescale them
func NormalizeCategory(userCategories []string) map[string]float64 {
	normalized := make(map[string]float64)
	log.Printf("🔎 Mapping categories: %v", userCategories)

	for _, movieCategory := range userCategories {
		if movieCategory == "" {
			log.Println("⚠️ Skipping empty category")
			continue
		}

		mappedAdCategories, err := FetchMappedAdCategories(movieCategory)
		if err != nil {
			log.Printf("❌ Error fetching mapped categories for %s: %v", movieCategory, err)
			continue
		}

		for adCategory, weight := range mappedAdCategories {
			normalized[adCategory] = (normalized[adCategory] + weight) / 2
		}
	}

	// Normalize the category scores
	var maxVal, minVal float64 = math.Inf(-1), math.Inf(1)
	for _, v := range normalized {
		if v > maxVal {
			maxVal = v
		}
		if v < minVal {
			minVal = v
		}
	}

	if maxVal > minVal {
		for k, v := range normalized {
			normalized[k] = (v - minVal) / (maxVal - minVal)
		}
	}

	log.Printf("✅ Final Normalized Categories: %v", normalized)
	return normalized
}

// FetchAdsForRecommendation retrieves ads based on category weights
func FetchAdsForRecommendation(categories map[string]float64) ([]models.Ad, map[string]float64, error) {
	ads := []models.Ad{}
	seenAds := make(map[string]bool)
	categoryWeights := make(map[string]float64)

	log.Printf("🔍 Fetching ads for mapped categories: %v", categories)

	for category, weight := range categories {
		log.Printf("🔍 Querying AdTable for category: %s (weight: %.4f)", category, weight)

		output, err := db.DynamoClient.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName:        aws.String(db.AdTableName),
			FilterExpression: aws.String("category = :category"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":category": &types.AttributeValueMemberS{Value: category},
			},
		})

		if err != nil {
			log.Printf("❌ Failed to fetch ads for category %s: %v", category, err)
			continue
		}

		for _, item := range output.Items {
			adID := item["ad_id"].(*types.AttributeValueMemberS).Value

			if seenAds[adID] {
				log.Printf("⚠️ Skipping duplicate ad: %s", adID)
				continue
			}
			seenAds[adID] = true

			ad := models.Ad{
				AdID:        adID,
				Category:    item["category"].(*types.AttributeValueMemberS).Value,
				Description: item["description"].(*types.AttributeValueMemberS).Value,
			}

			ads = append(ads, ad)
			categoryWeights[category] = weight
		}
	}

	log.Printf("✅ Total Unique Ads Retrieved: %d", len(ads))
	return ads, categoryWeights, nil
}

// RankAdsByHybridScoring ranks ads using a combination of category score and BERT similarity
func RankAdsByHybridScoring(ads []models.Ad, adEmbeddings [][]float64, userVector []float64, categoryWeights map[string]float64) []models.Ad {
	if len(ads) == 0 || len(adEmbeddings) == 0 {
		log.Println("⚠️ No ads or embeddings available for ranking")
		return nil
	}

	type ScoredAd struct {
		Ad    models.Ad
		Score float64
	}

	scores := []ScoredAd{}

	for i, ad := range ads {
		bertScore := CosineSimilarity(userVector, adEmbeddings[i])
		categoryScore, exists := categoryWeights[ad.Category]
		if !exists {
			categoryScore = 0.0
		}

		// Adjust weights to balance category score and BERT similarity
		finalScore := (0.4 * categoryScore) + (0.6 * bertScore)

		log.Printf("📊 AdID: %s, Category: %s, Category Score: %.4f, BERT Score: %.4f, Final Score: %.4f",
			ad.AdID, ad.Category, categoryScore, bertScore, finalScore)

		scores = append(scores, ScoredAd{Ad: ad, Score: finalScore})
	}

	// Sort Ads by Final Score (Descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Select Top 5 Ads
	topN := 5
	if len(scores) < topN {
		topN = len(scores)
	}

	rankedAds := make([]models.Ad, topN)
	for i := 0; i < topN; i++ {
		rankedAds[i] = scores[i].Ad
		log.Printf("🏆 Ranked Ad #%d - ID: %s, Final Score: %.4f", i+1, scores[i].Ad.AdID, scores[i].Score)
	}

	log.Printf("✅ Final Ranked Ads: %+v", rankedAds)
	return rankedAds
}

// GenerateRecommendations generates ad recommendations
func GenerateRecommendations(userID string) []models.Ad {
	log.Printf("🔍 Generating recommendations for user: %s", userID)

	// Fetch user playback history
	playbackHistory, err := FetchUserPlaybackHistory(userID)
	if err != nil {
		log.Printf("❌ Failed to fetch playback history: %v", err)
		return nil
	}

	// Normalize category scores
	mappedCategories := NormalizeCategory(playbackHistory)

	// Fetch ads based on mapped categories
	ads, categoryWeights, err := FetchAdsForRecommendation(mappedCategories)
	if err != nil || len(ads) == 0 {
		log.Println("⚠️ No ads found for mapped categories")
		return nil
	}

	// Extract ad descriptions for BERT embeddings
	adTexts := make([]string, len(ads))
	for i, ad := range ads {
		adTexts[i] = ad.Description
	}

	// Generate BERT embeddings for ads
	adEmbeddings, err := GenerateBERTEmbeddings(adTexts)
	if err != nil {
		log.Println("❌ Failed to generate BERT embeddings for ads")
		return nil
	}

	// Generate BERT embeddings for user's playback history
	historyEmbeddings, err := GenerateBERTEmbeddings(playbackHistory)
	if err != nil {
		log.Println("❌ Failed to generate BERT embeddings for user history")
		return nil
	}

	// Compute user embedding vector
	userVector := ComputeUserVector(historyEmbeddings)
	log.Printf("📊 Computed User Embedding Vector")

	// Rank Ads using Hybrid Scoring (Category + BERT Scores)
	rankedAds := RankAdsByHybridScoring(ads, adEmbeddings, userVector, categoryWeights)

	return rankedAds
}

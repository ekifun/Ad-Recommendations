package services

import (
	"Ad-Recommendations/db"
	"Ad-Recommendations/models"
	"context"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Map of mapped categories to actual AdTable categories
var CategoryNormalizationMap = map[string]string{
	"Gadgets":            "Tech",
	"Software":           "Tech",
	"Airlines":           "Travel",
	"Hotels":             "Travel",
	"Gym Equipment":      "Fitness",
	"Health Supplements": "Fitness",
	"Events":             "Entertainment",
	"Streaming Services": "Entertainment",
}

// Normalize categories before querying AdTable
func NormalizeCategory(mappedCategories []string) []string {
	normalized := make(map[string]bool)
	for _, category := range mappedCategories {
		if norm, exists := CategoryNormalizationMap[category]; exists {
			normalized[norm] = true
		}
	}

	// Convert to list
	result := []string{}
	for cat := range normalized {
		result = append(result, cat)
	}
	return result
}

// FetchUserPlaybackHistory retrieves the playback history for a user from DynamoDB
func FetchUserPlaybackHistory(userID string) ([]string, error) {
	output, err := db.DynamoClient.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(db.PlaybackTableName),
		KeyConditionExpression: aws.String("user_id = :userID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		log.Printf("Failed to query playback history: %v", err)
		return nil, err
	}

	playbackHistory := []string{}
	for _, item := range output.Items {
		if category, ok := item["category"].(*types.AttributeValueMemberS); ok {
			playbackHistory = append(playbackHistory, category.Value)
		}
	}

	return playbackHistory, nil
}

// FetchUserAdClickHistory retrieves the ad click history for a user from DynamoDB
func FetchUserAdClickHistory(userID string) ([]string, error) {
	output, err := db.DynamoClient.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(db.AdClickTableName),
		KeyConditionExpression: aws.String("user_id = :userID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		log.Printf("Failed to query ad click history: %v", err)
		return nil, err
	}

	adClickHistory := []string{}
	for _, item := range output.Items {
		if adID, ok := item["ad_id"].(*types.AttributeValueMemberS); ok {
			adClickHistory = append(adClickHistory, adID.Value)
		}
	}

	return adClickHistory, nil
}

// FetchMappedAdCategories retrieves the mapped ad categories for a given movie category
func FetchMappedAdCategories(movieCategory string) ([]string, error) {
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
		return []string{}, nil
	}

	mappedCategories := []string{}
	if categoriesAttr, ok := output.Item["ad_categories"].(*types.AttributeValueMemberSS); ok {
		mappedCategories = categoriesAttr.Value
	}

	log.Printf("✅ Mapped categories for %s: %v", movieCategory, mappedCategories)
	return mappedCategories, nil
}

// FetchAdsForRecommendation retrieves ads from DynamoDB based on mapped categories
func FetchAdsForRecommendation(categories []string) ([]models.Ad, error) {
	ads := []models.Ad{}
	seenAds := make(map[string]bool) // Track seen ads to remove duplicates

	for _, category := range categories {
		log.Printf("🔍 Querying AdTable for category: %s", category)

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

			// Skip duplicate ads
			if seenAds[adID] {
				continue
			}
			seenAds[adID] = true

			ads = append(ads, models.Ad{
				AdID:        adID,
				Category:    item["category"].(*types.AttributeValueMemberS).Value,
				Description: item["description"].(*types.AttributeValueMemberS).Value,
			})
		}
	}

	log.Printf("✅ Total Unique Ads Retrieved: %d", len(ads))
	return ads, nil
}

// padEmbeddings ensures the embeddings list matches the number of ads
func padEmbeddings(embeddings [][]float64, requiredSize int) [][]float64 {
	currentSize := len(embeddings)
	embeddingLength := len(embeddings[0]) // Assumes all embeddings are the same length

	for i := currentSize; i < requiredSize; i++ {
		zeroVector := make([]float64, embeddingLength)
		embeddings = append(embeddings, zeroVector)
	}

	log.Printf("✅ Adjusted embeddings to match ads: %d embeddings for %d ads", len(embeddings), requiredSize)
	return embeddings
}

// getMappedCategories converts movie categories to mapped ad categories
func getMappedCategories(movieCategories []string) []string {
	mappedCategories := make(map[string]bool)

	for _, category := range movieCategories {
		log.Printf("🔍 Fetching mapped ad categories for movie category: %s", category)
		adCategories, err := FetchMappedAdCategories(category)
		if err != nil {
			log.Printf("❌ Error fetching mapped ad categories for %s: %v", category, err)
			continue
		}

		for _, adCategory := range adCategories {
			mappedCategories[adCategory] = true // Avoid duplicates
		}
	}

	// Convert map keys to slice
	result := make([]string, 0, len(mappedCategories))
	for key := range mappedCategories {
		result = append(result, key)
	}

	log.Printf("✅ Mapped ad categories: %v", result)
	return result
}

// RankAdsBySimilarity ranks ads based on user similarity scores
// RankAdsBySimilarity ranks ads based on similarity and removes duplicates
func RankAdsBySimilarity(ads []models.Ad, embeddings [][]float64, userVector []float64) []models.Ad {
	if len(ads) == 0 || len(embeddings) == 0 {
		log.Println("⚠️ No ads or embeddings available for ranking")
		return nil
	}

	scores := make([]struct {
		Ad    models.Ad
		Score float64
	}, len(ads))

	seenAds := make(map[string]bool) // Track unique ads

	for i, ad := range ads {
		if i >= len(embeddings) {
			log.Printf("⚠️ Skipping ad %s due to missing embedding", ad.AdID)
			continue
		}

		if seenAds[ad.AdID] {
			continue // Skip duplicate ads
		}
		seenAds[ad.AdID] = true

		scores[i] = struct {
			Ad    models.Ad
			Score float64
		}{
			Ad:    ad,
			Score: CosineSimilarity(userVector, embeddings[i]),
		}
	}

	// Sort ads by score
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Select the top N recommendations
	topN := 5
	if len(scores) < topN {
		topN = len(scores)
	}

	rankedAds := make([]models.Ad, topN)
	for i := 0; i < topN; i++ {
		rankedAds[i] = scores[i].Ad
	}

	log.Printf("✅ Final Ranked Unique Ads: %+v", rankedAds)
	return rankedAds
}

// GenerateRecommendations generates ad recommendations based on user history
// GenerateRecommendations ensures ads and embeddings are correctly matched before processing
func GenerateRecommendations(userID string, embeddings [][]float64) []models.Ad {
	log.Printf("🔍 Generating recommendations for user: %s", userID)

	// Fetch user playback and ad click history
	playbackHistory, err := FetchUserPlaybackHistory(userID)
	if err != nil {
		log.Printf("❌ Failed to fetch playback history: %v", err)
		return nil
	}
	adClickHistory, err := FetchUserAdClickHistory(userID)
	if err != nil {
		log.Printf("❌ Failed to fetch ad click history: %v", err)
		return nil
	}

	// Fetch recommended ads
	mappedCategories := getMappedCategories(playbackHistory) // Converts movie categories to ad categories
	ads, err := FetchAdsForRecommendation(mappedCategories)
	if err != nil || len(ads) == 0 {
		log.Println("⚠️ No ads found for mapped categories")
		return nil
	}

	log.Printf("✅ Total Ads Retrieved: %d", len(ads))

	// Ensure embeddings match the number of ads
	if len(embeddings) < len(ads) {
		log.Printf("⚠️ Warning: Not enough embeddings (%d) for ads (%d). Adjusting...", len(embeddings), len(ads))
		embeddings = padEmbeddings(embeddings, len(ads))
	}

	// Generate user vector
	userVector := GenerateUserVector(playbackHistory, adClickHistory, ads, embeddings)

	// Rank ads based on similarity
	return RankAdsBySimilarity(ads, embeddings, userVector)
}

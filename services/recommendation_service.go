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
		} else {
			log.Printf("Invalid category data in playback history for user: %s", userID)
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
		} else {
			log.Printf("Invalid ad_id data in ad click history for user: %s", userID)
		}
	}

	return adClickHistory, nil
}

// GenerateRecommendations generates ad recommendations based on user history and embeddings
func GenerateRecommendations(playbackHistory, adClickHistory []string, ads []models.Ad, embeddings [][]float64) []models.Ad {
	// Generate a user vector based on playback and ad click history
	userVector := GenerateUserVector(playbackHistory, adClickHistory, ads, embeddings)

	// Calculate similarity scores for each ad
	scores := []struct {
		Ad    models.Ad
		Score float64
	}{}

	for i, embedding := range embeddings {
		score := CosineSimilarity(userVector, embedding)
		scores = append(scores, struct {
			Ad    models.Ad
			Score float64
		}{ads[i], score})
	}

	// Sort ads by score in descending order
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Select the top 5 recommendations
	recommendations := []models.Ad{}
	for i := 0; i < len(scores) && i < 5; i++ {
		recommendations = append(recommendations, scores[i].Ad)
	}

	return recommendations
}

package services

import (
	"Ad-Recommendations/db"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// AdClickEntry represents an ad click event
type AdClickEntry struct {
	UserID    string `json:"user_id"`
	AdID      string `json:"ad_id"`
	Timestamp string `json:"timestamp"`
}

// LogAdClick logs an ad click event to the DynamoDB AdClickTable
func LogAdClick(userID, adID string) error {
	// Validate inputs before logging
	if userID == "" {
		return fmt.Errorf("user_id cannot be empty")
	}
	if adID == "" {
		return fmt.Errorf("ad_id cannot be empty")
	}

	// Get the current timestamp
	timestamp := time.Now().Format(time.RFC3339)

	// Create the entry to insert
	entry := map[string]types.AttributeValue{
		"user_id":   &types.AttributeValueMemberS{Value: userID},
		"ad_id":     &types.AttributeValueMemberS{Value: adID},
		"timestamp": &types.AttributeValueMemberS{Value: timestamp},
	}

	log.Printf("Logging Ad Click: user_id=%s, ad_id=%s", userID, adID)

	// Perform the PutItem operation
	_, err := db.DynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(db.AdClickTableName),
		Item:      entry,
	})
	if err != nil {
		log.Printf("Failed to log ad click event: %v", err)
		return err
	}

	log.Printf("Ad click event logged: UserID=%s, AdID=%s, Timestamp=%s", userID, adID, timestamp)
	return nil
}

// GetAdClickHistory retrieves the ad click history for a given user
func GetAdClickHistory(userID string) ([]AdClickEntry, error) {
	// Validate input
	if userID == "" {
		return nil, fmt.Errorf("user_id cannot be empty")
	}

	// Query the AdClickTable for entries matching the user_id
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

	// Handle empty results
	if len(output.Items) == 0 {
		log.Printf("No ad click history found for user_id=%s", userID)
		return nil, nil
	}

	// Parse the query results into a list of AdClickEntry
	adClickHistory := []AdClickEntry{}
	for _, item := range output.Items {
		entry := AdClickEntry{
			UserID:    item["user_id"].(*types.AttributeValueMemberS).Value,
			AdID:      item["ad_id"].(*types.AttributeValueMemberS).Value,
			Timestamp: item["timestamp"].(*types.AttributeValueMemberS).Value,
		}
		adClickHistory = append(adClickHistory, entry)
	}

	return adClickHistory, nil
}

package services

import (
	"Ad-Recommendations/db"
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// LogPlayback logs playback data to DynamoDB
func LogPlayback(userID, category string) error {
	timestamp := time.Now().Format(time.RFC3339)

	entry := map[string]types.AttributeValue{
		"user_id":   &types.AttributeValueMemberS{Value: userID},
		"category":  &types.AttributeValueMemberS{Value: category},
		"timestamp": &types.AttributeValueMemberS{Value: timestamp},
	}

	_, err := db.DynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(db.PlaybackTableName),
		Item:      entry,
	})
	if err != nil {
		log.Printf("Failed to log playback data: %v", err)
		return err
	}

	log.Printf("Playback data logged: UserID=%s, Category=%s, Timestamp=%s", userID, category, timestamp)
	return nil
}

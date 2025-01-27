package services

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Interaction struct {
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	Target    string `json:"target"`
	Timestamp string `json:"timestamp"`
}

type InteractionService struct {
	DynamoClient         *dynamodb.Client
	InteractionTableName string
}

func (s *InteractionService) LogInteraction(userID, interactionType, target string) error {
	entry := map[string]types.AttributeValue{
		"user_id":   &types.AttributeValueMemberS{Value: userID},
		"type":      &types.AttributeValueMemberS{Value: interactionType},
		"target":    &types.AttributeValueMemberS{Value: target},
		"timestamp": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
	}

	_, err := s.DynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &s.InteractionTableName,
		Item:      entry,
	})
	if err != nil {
		log.Printf("Failed to log interaction: %v", err)
		return err
	}

	log.Printf("Logged interaction: userID=%s, type=%s, target=%s", userID, interactionType, target)
	return nil
}

func (s *InteractionService) FetchInteractions(userID string) ([]Interaction, error) {
	output, err := s.DynamoClient.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              &s.InteractionTableName,
		KeyConditionExpression: aws.String("user_id = :userID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		log.Printf("Failed to fetch interactions: %v", err)
		return nil, err
	}

	interactions := []Interaction{}
	for _, item := range output.Items {
		interaction := Interaction{
			UserID:    item["user_id"].(*types.AttributeValueMemberS).Value,
			Type:      item["type"].(*types.AttributeValueMemberS).Value,
			Target:    item["target"].(*types.AttributeValueMemberS).Value,
			Timestamp: item["timestamp"].(*types.AttributeValueMemberS).Value,
		}
		interactions = append(interactions, interaction)
	}

	return interactions, nil
}

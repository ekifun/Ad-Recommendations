package services

import (
	"Ad-Recommendations/db"
	"Ad-Recommendations/models"
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// AddUser adds a new user to the Users table in DynamoDB
func AddUser(user models.User) error {
	// Validate input
	if user.UserID == "" {
		return errors.New("user_id cannot be empty")
	}

	// Prepare PutItem input
	_, err := db.DynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(db.UserTableName),
		Item:      user.ToDynamoDBItem(),
	})

	if err != nil {
		return errors.New("failed to add user: " + err.Error())
	}

	return nil
}

// GetUser retrieves a user from the Users table in DynamoDB
// GetUser retrieves a user from DynamoDB
func GetUser(userID string) (*models.User, error) {
	// Validate input
	if userID == "" {
		return nil, errors.New("user_id cannot be empty")
	}

	// Query DynamoDB
	output, err := db.DynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(db.UserTableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
		},
	})

	if err != nil {
		return nil, errors.New("failed to fetch user: " + err.Error())
	}

	// Check if user exists
	if output.Item == nil {
		return nil, errors.New("user not found")
	}

	// Convert DynamoDB item to User model
	user := models.UserFromDynamoDBItem(output.Item)

	return user, nil
}

// UpdateUser updates an existing user in the Users table
func UpdateUser(user models.User) error {
	// Reuse AddUser for the full replacement
	return AddUser(user)
}

// DeleteUser removes a user from the Users table in DynamoDB
func DeleteUser(userID string) error {
	// Validate input
	if userID == "" {
		return errors.New("user_id cannot be empty")
	}

	// Delete the user
	_, err := db.DynamoClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(db.UserTableName),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
		},
	})

	if err != nil {
		return errors.New("failed to delete user: " + err.Error())
	}

	return nil
}

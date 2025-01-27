package models

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// User represents a user profile
type User struct {
	UserID  string   `json:"userID"`  // Match the JSON input field
	History []string `json:"history"` // User interaction history
}

// ToDynamoDBItem converts a User object to a DynamoDB item
func (u *User) ToDynamoDBItem() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"user_id": &types.AttributeValueMemberS{Value: u.UserID},
		"history": &types.AttributeValueMemberSS{Value: u.History},
	}
}

// UserFromDynamoDBItem converts a DynamoDB item to a User struct
func UserFromDynamoDBItem(item map[string]types.AttributeValue) *User {
	user := &User{
		UserID:  item["user_id"].(*types.AttributeValueMemberS).Value,
		History: []string{},
	}

	if history, ok := item["history"].(*types.AttributeValueMemberSS); ok {
		user.History = history.Value
	}

	return user
}

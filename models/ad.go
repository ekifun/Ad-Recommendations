package models

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Ad represents an advertisement
type Ad struct {
	AdID        string   `json:"ad_id"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

// FetchAllAds retrieves all ads from the DynamoDB Ads table
func FetchAllAds(ctx context.Context, dynamoClient *dynamodb.Client, tableName string) ([]Ad, error) {
	// Validate the table name
	if tableName == "" {
		return nil, fmt.Errorf("tableName cannot be empty")
	}

	// Scan the Ads table
	output, err := dynamoClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: &tableName,
	})
	if err != nil {
		log.Printf("Failed to scan Ads table: %v", err)
		return nil, fmt.Errorf("failed to scan Ads table: %w", err)
	}

	ads := []Ad{}

	// Parse the results
	for _, item := range output.Items {
		ad := Ad{}

		// Safely extract attributes
		if adID, ok := item["ad_id"].(*types.AttributeValueMemberS); ok {
			ad.AdID = adID.Value
		}

		if category, ok := item["category"].(*types.AttributeValueMemberS); ok {
			ad.Category = category.Value
		}

		if description, ok := item["description"].(*types.AttributeValueMemberS); ok {
			ad.Description = description.Value
		}

		if keywords, ok := item["keywords"].(*types.AttributeValueMemberSS); ok {
			ad.Keywords = keywords.Value
		}

		ads = append(ads, ad)
	}

	return ads, nil
}

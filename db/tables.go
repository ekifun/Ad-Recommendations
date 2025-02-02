package db

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	UserTableName            = "Users"
	PlaybackTableName        = "PlaybackTable"
	AdClickTableName         = "AdClickTable"
	CategoryMappingTableName = "CategoryMappingTable" // ✅ Define Category Mapping Table
	AdTableName              = "AdTable"              // ✅ Define Ad Table
)

// EnsureTables ensures the existence of required tables in DynamoDB
func EnsureTables() error {
	tables := []struct {
		Name          string
		KeySchema     []types.KeySchemaElement
		AttributeDefs []types.AttributeDefinition
	}{
		{
			Name: UserTableName,
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
			},
			AttributeDefs: []types.AttributeDefinition{
				{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
			},
		},
		{
			Name: PlaybackTableName,
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
			},
			AttributeDefs: []types.AttributeDefinition{
				{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
			},
		},
		{
			Name: AdClickTableName,
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("user_id"), KeyType: types.KeyTypeHash},
				{AttributeName: aws.String("ad_id"), KeyType: types.KeyTypeRange},
			},
			AttributeDefs: []types.AttributeDefinition{
				{AttributeName: aws.String("user_id"), AttributeType: types.ScalarAttributeTypeS},
				{AttributeName: aws.String("ad_id"), AttributeType: types.ScalarAttributeTypeS},
			},
		},
		{
			Name: CategoryMappingTableName, // ✅ Ensure Category Mapping Table
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("movie_category"), KeyType: types.KeyTypeHash},
			},
			AttributeDefs: []types.AttributeDefinition{
				{AttributeName: aws.String("movie_category"), AttributeType: types.ScalarAttributeTypeS},
			},
		},
		{
			Name: AdTableName, // ✅ Ensure Ad Table
			KeySchema: []types.KeySchemaElement{
				{AttributeName: aws.String("ad_id"), KeyType: types.KeyTypeHash},
			},
			AttributeDefs: []types.AttributeDefinition{
				{AttributeName: aws.String("ad_id"), AttributeType: types.ScalarAttributeTypeS},
			},
		},
	}

	for _, table := range tables {
		log.Printf("Ensuring table %s exists", table.Name)

		_, err := DynamoClient.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
			TableName:            aws.String(table.Name),
			KeySchema:            table.KeySchema,
			AttributeDefinitions: table.AttributeDefs,
			BillingMode:          types.BillingModePayPerRequest,
		})

		if err != nil {
			log.Printf("Table %s might already exist: %v", table.Name, err)
		} else {
			log.Printf("Table %s created successfully", table.Name)
		}

		if err := WaitUntilTableActive(table.Name); err != nil {
			log.Printf("Error waiting for table %s to become active: %v", table.Name, err)
			return err
		}
	}

	return nil
}

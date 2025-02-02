package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var DynamoClient *dynamodb.Client

// InitDynamoDB initializes the DynamoDB client
func InitDynamoDB() {
	// Load the AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		)),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS configuration: %v", err)
	}

	// Create the DynamoDB client
	DynamoClient = dynamodb.NewFromConfig(cfg)
	log.Println("DynamoDB client initialized")

	// Ensure tables exist
	if err := EnsureTables(); err != nil {
		log.Fatalf("Failed to ensure tables exist: %v", err)
	}
}

// WaitUntilTableActive waits until a DynamoDB table becomes active
func WaitUntilTableActive(tableName string) error {
	for {
		describeOutput, err := DynamoClient.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return err
		}

		if describeOutput.Table.TableStatus == "ACTIVE" {
			log.Printf("Table %s is active", tableName)
			return nil
		}

		time.Sleep(2 * time.Second)
	}
}

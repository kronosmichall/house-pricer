package db

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"house-pricer/types"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const batch_size = 25

type Client struct {
	db        *dynamodb.Client
	tableName string
	ctx       context.Context
}

func GetClient() Client {
	ctx := context.Background()
	region := os.Getenv("AWS_REGION")
	if region == "" {
		panic(errors.New("AWS region not set"))
	}
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		panic(errors.New("DynamoDB table name not set"))
	}

	cnf, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		panic(err)
	}

	db := dynamodb.NewFromConfig(cnf)
	return Client{db, tableName, ctx}
}

func Insert(client Client, offerts []types.Offert) error {
	for i := 0; i < len(offerts); i++ {
		end := i + batch_size
		if end > len(offerts) {
			end = len(offerts)
		}
		batch := offerts[i:end]
		err := insertBatch(client, batch)
		if err != nil {
			return err
		}
	}

	return nil
}

func insertBatch(client Client, offerts []types.Offert) error {
	writeRequests := make([]dbTypes.WriteRequest, 0, len(offerts))
	now := time.Now()

	for _, offert := range offerts {
		offert.CreatedAt = now
		offert.LastUpdated = now

		av, err := attributevalue.MarshalMap(offert)
		if err != nil {
			log.Println("Failed to marshal item ", offert)
			return err
		}
		writeRequest := dbTypes.WriteRequest{
			PutRequest: &dbTypes.PutRequest{
				Item: av,
			},
		}
		writeRequests = append(writeRequests, writeRequest)
	}

	if len(writeRequests) == 0 {
		log.Println("No valid items to write in this batch, skipping.")
		return nil
	}

	batchInput := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]dbTypes.WriteRequest{
			"tableName": writeRequests,
		},
	}

	_, err := client.db.BatchWriteItem(client.ctx, batchInput)

	return err
}

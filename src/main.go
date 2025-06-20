package main

import (
	"context"

	"house-pricer/db"
	"house-pricer/scraper"
	"house-pricer/types"

	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context) (string, error) {
	resultChannel := make(chan []types.Offert)
	go scraper.FetchAll(resultChannel)
	client := db.GetClient()
	err := db.Insert(client, resultChannel)

	return "Hi", err
}

func main() {
	lambda.Start(HandleRequest)
}

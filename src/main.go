package main

import (
	"context"
	"house-pricer/db"
	"house-pricer/scraper"

	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context) (string, error) {
	offerts := scraper.FetchAll()
	client := db.GetClient()
	err := db.Insert(client, offerts)

	return "Hi", err
}

func main() {
	lambda.Start(HandleRequest)
}

package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context) (string, error) {
	offerts := fetchAll()
	return "Hello", nil
}

func main() {
	lambda.Start(HandleRequest)
}

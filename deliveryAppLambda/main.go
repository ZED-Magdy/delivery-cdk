package main

import (
	"github.com/ZED-Magdy/delivery-cdk/lambda/handlers"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)


func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch {
	case request.Path == "/ads":
		return handlers.GetAds(request)
	case request.Path == "/categories":
		return handlers.GetCategories(request)
	case request.Resource == "/products/{categoryId}" || pathStartsWith(request.Path, "/products/"):
		return handlers.GetProducts(request)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "Not Found",
		}, nil
	}
}

func pathStartsWith(path, prefix string) bool {
	return len(path) >= len(prefix) && path[:len(prefix)] == prefix
}

func main() {
	lambda.Start(handler)
}
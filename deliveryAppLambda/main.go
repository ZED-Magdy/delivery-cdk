package main

import (
	"github.com/ZED-Magdy/delivery-cdk/lambda/handlers"
	"github.com/ZED-Magdy/delivery-cdk/lambda/middlewares"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch {
	case request.Path == "/ads":
		return middlewares.AuthMiddleware(handlers.GetAds)(request)
	case request.Path == "/categories":
		return middlewares.AuthMiddleware(handlers.GetCategories)(request)
	case request.Resource == "/products/{categoryId}" || pathStartsWith(request.Path, "/products/"):
		return middlewares.AuthMiddleware(handlers.GetProducts)(request)
	case request.Path == "/users/register" && request.HTTPMethod == "POST":
		return handlers.RegisterUser(request)
	case request.Path == "/users/send-otp" && request.HTTPMethod == "POST":
		return handlers.SendOTP(request)
	case request.Path == "/users/verify-otp" && request.HTTPMethod == "POST":
		return handlers.VerifyOTP(request)
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
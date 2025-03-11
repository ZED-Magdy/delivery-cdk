package main

import (
	"github.com/ZED-Magdy/delivery-cdk/lambda/handlers"
	"github.com/ZED-Magdy/delivery-cdk/lambda/middlewares"
	"github.com/ZED-Magdy/delivery-cdk/lambda/router"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	router := setupRouter()
	
	handler, found := router.Match(request)
	if !found {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "Not Found",
		}, nil
	}
	
	return handler(request)
}

func setupRouter() *router.Router {
	r := router.NewRouter()
	
	authMiddleware := middlewares.AdaptAuthMiddleware()
	r.Add("/users/register", "POST", handlers.RegisterUser)
	r.Add("/users/send-otp", "POST", handlers.SendOTP)
	r.Add("/users/verify-otp", "POST", handlers.VerifyOTP)
	r.Add("/ads", "GET", handlers.GetAds, authMiddleware)
	r.Add("/categories", "GET", handlers.GetCategories, authMiddleware)
	r.Add("/products/{categoryId}", "GET", handlers.GetProducts, authMiddleware)
	
	
	return r
}

func main() {
	lambda.Start(handler)
}
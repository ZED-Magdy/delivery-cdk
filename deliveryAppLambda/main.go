package main

import (
	"github.com/ZED-Magdy/delivery-cdk/lambda/handlers"
	"github.com/ZED-Magdy/delivery-cdk/lambda/middlewares"
	"github.com/ZED-Magdy/delivery-cdk/lambda/router"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func setupRouter() *router.Router {
	r := router.NewRouter()
	
	authMiddleware := middlewares.AdaptAuthMiddleware()

	r.Add("/users/register", "POST", handlers.RegisterUser)
	r.Add("/users/send-otp", "POST", handlers.SendOTP)
	r.Add("/users/verify-otp", "POST", handlers.VerifyOTP)
	r.Add("/ads", "GET", handlers.GetAds, authMiddleware)
	r.Add("/categories", "GET", handlers.GetCategories, authMiddleware)
	r.Add("/products/{categoryId}", "GET", handlers.GetProducts, authMiddleware)
	r.Add("/orders", "POST", handlers.CreateOrder, authMiddleware)
	r.Add("/orders", "GET", handlers.GetUserOrders, authMiddleware)
	r.Add("/orders/{orderId}", "GET", handlers.GetOrderDetails, authMiddleware)
	r.Add("/orders/{orderId}/cancel", "POST", handlers.CancelOrder, authMiddleware)
	
	return r
}

func main() {
	lambda.Start(func (request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	router := setupRouter()
	
	handler, found := router.Match(request)
	if !found {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "Not Found",
		}, nil
	}
	
	return handler(request)
})
}
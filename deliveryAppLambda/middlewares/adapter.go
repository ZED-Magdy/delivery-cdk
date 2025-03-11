package middlewares

import (
	"github.com/ZED-Magdy/delivery-cdk/lambda/router"
	"github.com/aws/aws-lambda-go/events"
)

func AdaptAuthMiddleware() router.MiddlewareFunc {
	return func(next router.RouteHandler) router.RouteHandler {
		adaptedNext := func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return next(req)
		}
		
		authWrappedHandler := AuthMiddleware(adaptedNext)
		
		return func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return authWrappedHandler(req)
		}
	}
}

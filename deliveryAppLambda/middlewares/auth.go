package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ZED-Magdy/delivery-cdk/lambda/utils"
	"github.com/aws/aws-lambda-go/events"
)

type HandlerFunc func(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

func AuthMiddleware(handlerFunc HandlerFunc) HandlerFunc {
	return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		// Extract token from Authorization header
		authHeader := request.Headers["Authorization"]
		if authHeader == "" {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusUnauthorized,
				Body:       "Authorization header is required",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			}, nil
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusUnauthorized,
				Body:       "Authorization header format must be Bearer {token}",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			}, nil
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			errorResponse := map[string]string{"error": "Invalid token: " + err.Error()}
			jsonResponse, _ := json.Marshal(errorResponse)
			
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusUnauthorized,
				Body:       string(jsonResponse),
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			}, nil
		}

		request.Headers["X-User-ID"] = claims.UserID
		request.Headers["X-User-Name"] = claims.Name
		request.Headers["X-User-Phone"] = claims.Phone

		return handlerFunc(request)
	}
}

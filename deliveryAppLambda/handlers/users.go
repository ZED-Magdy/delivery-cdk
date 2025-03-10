package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ZED-Magdy/delivery-cdk/lambda/models"
	"github.com/aws/aws-lambda-go/events"
)

func RegisterUser(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var input models.UserRegistrationInput
	if err := json.Unmarshal([]byte(request.Body), &input); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request format",
		}, nil
	}

	if input.Name == "" || input.Phone == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Name and phone are required",
		}, nil
	}

	_, err := models.RegisterUser(input)
	if err != nil {
		if err.Error() == "phone number already registered" {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusConflict,
				Body:       "Phone number already registered",
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Error registering user: " + err.Error(),
		}, nil
	}

	

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       "{\"message\": \"User registered successfully\"}",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

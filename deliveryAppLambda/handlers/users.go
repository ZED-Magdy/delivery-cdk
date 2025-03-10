package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ZED-Magdy/delivery-cdk/lambda/models"
	"github.com/ZED-Magdy/delivery-cdk/lambda/utils"
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

func VerifyOTP(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var input models.OTPVerificationInput
	if err := json.Unmarshal([]byte(request.Body), &input); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request format",
		}, nil
	}

	if input.Phone == "" || input.OTP == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Phone and OTP are required",
		}, nil
	}

	user, err := models.VerifyOTP(input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Error verifying OTP: " + err.Error()

		switch err.Error() {
		case "user not found":
			statusCode = http.StatusNotFound
			message = "User not found"
		case "invalid OTP":
			statusCode = http.StatusUnauthorized
			message = "Invalid OTP"
		case "OTP expired":
			statusCode = http.StatusUnauthorized
			message = "OTP has expired"
		}

		return events.APIGatewayProxyResponse{
			StatusCode: statusCode,
			Body:       message,
		}, nil
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Name, user.Phone)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Error generating token",
		}, nil
	}

	// Prepare response with user data and token
	response := struct {
		User  *models.User `json:"user"`
		Token string       `json:"token"`
	}{
		User:  user,
		Token: token,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Error converting response to JSON",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(jsonResponse),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func SendOTP(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var input models.SendOTPInput
	if err := json.Unmarshal([]byte(request.Body), &input); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request format",
		}, nil
	}

	if input.Phone == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Phone number is required",
		}, nil
	}

	err := models.SendOTP(input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		message := "Error sending OTP: " + err.Error()

		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
			message = "User not found with the provided phone number"
		}

		return events.APIGatewayProxyResponse{
			StatusCode: statusCode,
			Body:       message,
		}, nil
	}

	response := map[string]string{
		"message": "OTP sent successfully",
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Error converting response to JSON",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(jsonResponse),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

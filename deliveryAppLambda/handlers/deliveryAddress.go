package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/ZED-Magdy/delivery-cdk/lambda/models"
	"github.com/aws/aws-lambda-go/events"
)

type CreateDeliveryAddressRequest struct {
	Name        string  `json:"name"`
	AddressLine string  `json:"addressLine"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
}

func CreateDeliveryAddress(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user, err := models.GetAuthUser(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "Unauthorized: " + err.Error(),
		}, nil
	}

	userId := user.ID

	var createReq CreateDeliveryAddressRequest
	err = json.Unmarshal([]byte(request.Body), &createReq)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid request format: " + err.Error(),
		}, nil
	}

	if createReq.Name == "" || createReq.AddressLine == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Name and address line are required",
		}, nil
	}

	address, err := models.CreateDeliveryAddress(models.DeliveryAddress{
		UserId:      userId,
		Name:        createReq.Name,
		AddressLine: createReq.AddressLine,
		Latitude:    createReq.Latitude,
		Longitude:   createReq.Longitude,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error creating delivery address: " + err.Error(),
		}, nil
	}

	jsonBody, err := json.Marshal(address)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error converting response to JSON",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       string(jsonBody),
	}, nil
}

func GetUserDeliveryAddresses(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user, err := models.GetAuthUser(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "Unauthorized: " + err.Error(),
		}, nil
	}

	userId := user.ID

	addresses, err := models.GetUserDeliveryAddresses(userId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Error retrieving delivery addresses: %v", err),
		}, nil
	}

	jsonBody, err := json.Marshal(addresses)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error converting response to JSON",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonBody),
	}, nil
}

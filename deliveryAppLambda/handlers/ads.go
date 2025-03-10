package handlers

import (
	"encoding/json"

	"github.com/ZED-Magdy/delivery-cdk/lambda/models"
	"github.com/aws/aws-lambda-go/events"
)

func GetAds(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ads, err := models.ListAll()
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       err.Error(),
		}, nil
	}

	jsonBody, err := json.Marshal(ads)
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
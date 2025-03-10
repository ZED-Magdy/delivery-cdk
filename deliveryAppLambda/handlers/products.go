package handlers

import (
	"encoding/json"

	"github.com/ZED-Magdy/delivery-cdk/lambda/models"
	"github.com/aws/aws-lambda-go/events"
)

func GetProducts(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	categoryId, ok := request.PathParameters["categoryId"]
	
	if !ok {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Missing categoryId",
		}, nil
	}

	products, err := models.ListAllProducts(string(categoryId))
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       err.Error(),
		}, nil
	}

	jsonBody, err := json.Marshal(products)
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

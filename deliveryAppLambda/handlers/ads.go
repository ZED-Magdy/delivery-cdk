package handlers

import (
	"context"
	"encoding/json"

	"github.com/ZED-Magdy/delivery-cdk/lambda/database"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func GetAds(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	adsTable := database.GetTables().AdsTable
	ddbClient, err := database.NewDynamoDBClient(adsTable)

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error connecting to database",
		}, nil
	}

	data, err := ddbClient.Client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: &ddbClient.Table,
	})

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       err.Error(),
		}, nil
	}

	jsonBody, err := json.Marshal(data.Items)
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
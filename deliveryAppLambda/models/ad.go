package models

import (
	"context"

	"github.com/ZED-Magdy/delivery-cdk/lambda/database"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Ad struct {
	Id         string `json:"id" dynamodbav:"id"`
	ImageUrl   string `json:"imageUrl" dynamodbav:"imageUrl"`
	Action     string `json:"action" dynamodbav:"action"`
	ActionType string `json:"actionType" dynamodbav:"actionType"`
}

func ListAll() ([]Ad, error) {
	adsTable := database.GetTables().AdsTable
	ddbClient, err := database.NewDynamoDBClient(adsTable)
	if err != nil {
		return nil, err
	}

	data, err := ddbClient.Client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: &ddbClient.Table,
	})
	if err != nil {
		return nil, err
	}

	var ads []Ad
	err = attributevalue.UnmarshalListOfMaps(data.Items, &ads)
	if err != nil {
		return nil, err
	}

	return ads, nil
}

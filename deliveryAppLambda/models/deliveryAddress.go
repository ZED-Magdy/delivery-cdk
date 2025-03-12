package models

import (
	"context"
	"fmt"

	"github.com/ZED-Magdy/delivery-cdk/lambda/database"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DeliveryAddress struct {
	Id           string  `json:"id" dynamodbav:"id"`
	UserId       string  `json:"userId" dynamodbav:"userId"`
	Name         string  `json:"name" dynamodbav:"name"`
	AddressLine string  `json:"addressLine" dynamodbav:"addressLine"`
	Latitude     float64 `json:"latitude,omitempty" dynamodbav:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty" dynamodbav:"longitude,omitempty"`
}

func GetDeliveryAddressById(addressId string) (*DeliveryAddress, error) {
	deliveryAddressTable := database.GetTables().DeliverAddressTable
	ddbClient, err := database.NewDynamoDBClient(deliveryAddressTable)
	if err != nil {
		return nil, err
	}

	result, err := ddbClient.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &ddbClient.Table,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: addressId},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, fmt.Errorf("delivery address not found")
	}

	var address DeliveryAddress
	err = attributevalue.UnmarshalMap(result.Item, &address)
	if err != nil {
		return nil, err
	}

	return &address, nil
}

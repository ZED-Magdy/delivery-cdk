package models

import (
	"context"
	"fmt"

	"github.com/ZED-Magdy/delivery-cdk/lambda/database"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type DeliveryAddress struct {
	Id           string  `json:"id" dynamodbav:"id"`
	UserId       string  `json:"userId" dynamodbav:"userId"`
	Name         string  `json:"name" dynamodbav:"name"`
	AddressLine  string  `json:"addressLine" dynamodbav:"addressLine"`
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

func CreateDeliveryAddress(address DeliveryAddress) (*DeliveryAddress, error) {
	deliveryAddressTable := database.GetTables().DeliverAddressTable
	ddbClient, err := database.NewDynamoDBClient(deliveryAddressTable)
	if err != nil {
		return nil, err
	}

	if address.Id == "" {
		address.Id = uuid.New().String()
	}

	item, err := attributevalue.MarshalMap(address)
	if err != nil {
		return nil, err
	}

	_, err = ddbClient.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &ddbClient.Table,
		Item:      item,
	})
	if err != nil {
		return nil, err
	}

	return &address, nil
}

func GetUserDeliveryAddresses(userId string) ([]DeliveryAddress, error) {
	deliveryAddressTable := database.GetTables().DeliverAddressTable
	ddbClient, err := database.NewDynamoDBClient(deliveryAddressTable)
	if err != nil {
		return nil, err
	}

	result, err := ddbClient.Client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: &ddbClient.Table,
		FilterExpression: aws.String("userId = :userId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: userId},
		},
	})
	if err != nil {
		return nil, err
	}

	var addresses []DeliveryAddress
	err = attributevalue.UnmarshalListOfMaps(result.Items, &addresses)
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

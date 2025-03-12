package models

import (
	"context"

	"github.com/ZED-Magdy/delivery-cdk/lambda/database"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// OrderItem represents an item in an order
type OrderItem struct {
	Id        string  `json:"id" dynamodbav:"id"`
	OrderId   string  `json:"orderId" dynamodbav:"orderId"`
	Name      string  `json:"name" dynamodbav:"name"`
	Price     float64 `json:"price" dynamodbav:"price"`
	ProductId string  `json:"productId" dynamodbav:"productId"`
	Quantity  int     `json:"quantity" dynamodbav:"quantity"`
}

// CreateOrderItem adds a new item to an order
func CreateOrderItem(orderItem OrderItem) (*OrderItem, error) {
	orderItemsTable := database.GetTables().OrderItemsTable
	ddbClient, err := database.NewDynamoDBClient(orderItemsTable)
	if err != nil {
		return nil, err
	}

	// Generate a new UUID for the item ID if not provided
	if orderItem.Id == "" {
		orderItem.Id = uuid.New().String()
	}

	item, err := attributevalue.MarshalMap(orderItem)
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

	return &orderItem, nil
}

// GetOrderItems retrieves all items for a specific order
func GetOrderItems(orderId string) ([]OrderItem, error) {
	orderItemsTable := database.GetTables().OrderItemsTable
	ddbClient, err := database.NewDynamoDBClient(orderItemsTable)
	if err != nil {
		return nil, err
	}

	result, err := ddbClient.Client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName:        &ddbClient.Table,
		FilterExpression: aws.String("orderId = :orderId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":orderId": &types.AttributeValueMemberS{Value: orderId},
		},
	})
	if err != nil {
		return nil, err
	}

	var orderItems []OrderItem
	err = attributevalue.UnmarshalListOfMaps(result.Items, &orderItems)
	if err != nil {
		return nil, err
	}

	return orderItems, nil
}

// DeleteOrderItems deletes all items for a specific order (used when canceling an order)
func DeleteOrderItems(orderId string) error {
	// First get all items for this order
	items, err := GetOrderItems(orderId)
	if err != nil {
		return err
	}

	orderItemsTable := database.GetTables().OrderItemsTable
	ddbClient, err := database.NewDynamoDBClient(orderItemsTable)
	if err != nil {
		return err
	}

	// Delete each item
	for _, item := range items {
		_, err = ddbClient.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
			TableName: &ddbClient.Table,
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: item.Id},
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

package models

import (
	"context"
	"fmt"
	"time"

	"github.com/ZED-Magdy/delivery-cdk/lambda/database"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusConfirmed  OrderStatus = "confirmed"
	StatusDelivering OrderStatus = "delivering"
	StatusDelivered  OrderStatus = "delivered"
	StatusCanceled   OrderStatus = "canceled"
)

// Order represents an order in the system
type Order struct {
	Id                string      `json:"id" dynamodbav:"id"`
	UserId            string      `json:"userId" dynamodbav:"userId"`
	Total             float64     `json:"total" dynamodbav:"total"`
	Status            OrderStatus `json:"status" dynamodbav:"status"`
	DeliveryAddressId string      `json:"deliveryAddressId" dynamodbav:"deliveryAddressId"`
	CreatedAt         string      `json:"createdAt" dynamodbav:"createdAt"`
}

// CreateOrder creates a new order in the database
func CreateOrder(order Order) (*Order, error) {
	ordersTable := database.GetTables().OrdersTable
	ddbClient, err := database.NewDynamoDBClient(ordersTable)
	if err != nil {
		return nil, err
	}

	// Generate a new UUID for the order ID if not provided
	if order.Id == "" {
		order.Id = uuid.New().String()
	}
	
	// Set initial status and creation time if not provided
	if order.Status == "" {
		order.Status = StatusPending
	}
	if order.CreatedAt == "" {
		order.CreatedAt = time.Now().Format(time.RFC3339)
	}

	item, err := attributevalue.MarshalMap(order)
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

	return &order, nil
}

// GetOrderById retrieves an order by its ID
func GetOrderById(orderId string) (*Order, error) {
	ordersTable := database.GetTables().OrdersTable
	ddbClient, err := database.NewDynamoDBClient(ordersTable)
	if err != nil {
		return nil, err
	}

	result, err := ddbClient.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &ddbClient.Table,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: orderId},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, fmt.Errorf("order not found")
	}

	var order Order
	err = attributevalue.UnmarshalMap(result.Item, &order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

// UpdateOrderStatus updates the status of an order
func UpdateOrderStatus(orderId string, status OrderStatus) (*Order, error) {
	ordersTable := database.GetTables().OrdersTable
	ddbClient, err := database.NewDynamoDBClient(ordersTable)
	if err != nil {
		return nil, err
	}

	// First check if the order exists and get its current status
	order, err := GetOrderById(orderId)
	if err != nil {
		return nil, err
	}
	
	// Prevent cancellation if the order is not pending
	if status == StatusCanceled && order.Status != StatusPending {
		return nil, fmt.Errorf("only pending orders can be canceled")
	}

	// Update the order status
	_, err = ddbClient.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: &ddbClient.Table,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: orderId},
		},
		UpdateExpression: aws.String("SET #status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: string(status)},
		},
	})
	if err != nil {
		return nil, err
	}

	// Return the updated order
	order.Status = status
	return order, nil
}

// GetUserOrders retrieves all orders for a specific user
func GetUserOrders(userId string) ([]Order, error) {
	ordersTable := database.GetTables().OrdersTable
	ddbClient, err := database.NewDynamoDBClient(ordersTable)
	if err != nil {
		return nil, err
	}

	// Scan the table and filter by userId
	// In a production environment, we would use a GSI on userId
	result, err := ddbClient.Client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName:        &ddbClient.Table,
		FilterExpression: aws.String("userId = :userId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: userId},
		},
	})
	if err != nil {
		return nil, err
	}

	var orders []Order
	err = attributevalue.UnmarshalListOfMaps(result.Items, &orders)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ZED-Magdy/delivery-cdk/lambda/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
)

type OrderNotification struct {
	OrderId    string `json:"orderId"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	CustomerID string `json:"customerId"`
}

func SendOrderStatusNotification(orderId, status, userId string) error {
	topicARN := os.Getenv("ORDER_STATUS_NOTIFICATION_TOPIC_ARN")
	if topicARN == "" {
		return fmt.Errorf("ORDER_STATUS_NOTIFICATION_TOPIC_ARN environment variable is not set")
	}

	// Get order details
	order, err := models.GetOrderById(orderId)
	if err != nil {
		return fmt.Errorf("failed to get order details: %v", err)
	}

	// Get user details for additional notification data if needed
	user, err := models.GetUserByID(userId)
	if err != nil {
		return fmt.Errorf("failed to get user details: %v", err)
	}

	message := fmt.Sprintf("Your order #%s has been updated to %s", orderId, status)
	
	notification := OrderNotification{
		OrderId:    order.Id,
		Status:     status,
		Message:    message,
		CustomerID: user.ID,
	}

	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %v", err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to load AWS SDK config: %v", err)
	}

	client := sns.NewFromConfig(cfg)
	_, err = client.Publish(context.TODO(), &sns.PublishInput{
		TopicArn: aws.String(topicARN),
		Message:  aws.String(string(notificationJSON)),
		Subject:  aws.String(fmt.Sprintf("Order Status Update: %s", status)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"OrderId": {
				DataType:    aws.String("String"),
				StringValue: aws.String(order.Id),
			},
			"Status": {
				DataType:    aws.String("String"),
				StringValue: aws.String(status),
			},
		},
	})
	
	if err != nil {
		return fmt.Errorf("failed to publish notification: %v", err)
	}

	return nil
}

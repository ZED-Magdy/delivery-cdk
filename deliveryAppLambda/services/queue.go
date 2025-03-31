package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type OrderMessage struct {
	OrderId string `json:"orderId"`
	Status  string `json:"status"`
	UserId  string `json:"userId"`
}

func SendOrderToQueue(orderId, status, userId string) error {
	queueURL := os.Getenv("ORDER_QUEUE_URL")
	if queueURL == "" {
		return fmt.Errorf("ORDER_QUEUE_URL environment variable is not set")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to load AWS SDK config: %v", err)
	}

	client := sqs.NewFromConfig(cfg)

	message := OrderMessage{
		OrderId: orderId,
		Status:  status,
		UserId:  userId,
	}

	messageBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal order message: %v", err)
	}

	_, err = client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(messageBody)),
	})
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %v", err)
	}

	return nil
}

func ProcessOrdersFromQueue() error {
	queueURL := os.Getenv("ORDER_QUEUE_URL")
	if queueURL == "" {
		return fmt.Errorf("ORDER_QUEUE_URL environment variable is not set")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to load AWS SDK config: %v", err)
	}

	client := sqs.NewFromConfig(cfg)

	result, err := client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
	})
	if err != nil {
		return fmt.Errorf("failed to receive messages from SQS: %v", err)
	}

	for _, message := range result.Messages {
		var orderMsg OrderMessage
		err := json.Unmarshal([]byte(*message.Body), &orderMsg)
		if err != nil {
			fmt.Printf("Failed to unmarshal message: %v\n", err)
			continue
		}

		// Process the order - update status and send notification
		fmt.Printf("Processing order %s with status %s\n", orderMsg.OrderId, orderMsg.Status)
		
		// Send notification about the order status
		err = SendOrderStatusNotification(orderMsg.OrderId, orderMsg.Status, orderMsg.UserId)
		if err != nil {
			fmt.Printf("Failed to send notification: %v\n", err)
			// Continue processing other messages even if notification fails
		}

		// Delete the message from the queue after processing
		_, err = client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(queueURL),
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			fmt.Printf("Failed to delete message: %v\n", err)
		}
	}

	return nil
}

func ProcessOrderFromMessage(messageBody string) error {
	var orderMsg OrderMessage
	err := json.Unmarshal([]byte(messageBody), &orderMsg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	// Process the order - update status and send notification
	fmt.Printf("Processing order %s with status %s\n", orderMsg.OrderId, orderMsg.Status)
	
	// Send notification about the order status
	return SendOrderStatusNotification(orderMsg.OrderId, orderMsg.Status, orderMsg.UserId)
}

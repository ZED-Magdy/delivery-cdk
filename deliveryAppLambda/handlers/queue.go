package handlers

import (
	"github.com/ZED-Magdy/delivery-cdk/lambda/services"
	"github.com/aws/aws-lambda-go/events"
)

func ProcessOrderQueue(request events.SQSEvent) error {
	for _, record := range request.Records {
		// Process each SQS message
		// The services.ProcessOrdersFromQueue function would typically handle the logic
		// But for direct SQS Lambda triggers, we can process here directly
		err := services.ProcessOrderFromMessage(record.Body)
		if err != nil {
			// Log the error but continue processing other messages
			// This prevents the entire batch from failing due to one message
			// You might want to implement a dead-letter queue for failed messages
			// in a production system
		}
	}
	return nil
}

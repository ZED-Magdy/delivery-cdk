package models

import (
	"context"

	"github.com/ZED-Magdy/delivery-cdk/lambda/database"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Product represents a product in the system
type Product struct {
	Id          string  `json:"id" dynamodbav:"id"`
	Name        string  `json:"name" dynamodbav:"name"`
	Description string  `json:"description" dynamodbav:"description"`
	Price       float64 `json:"price" dynamodbav:"price"`
	ImageUrl    string  `json:"imageUrl" dynamodbav:"imageUrl"`
	CategoryId  string  `json:"categoryId" dynamodbav:"categoryId"`
}

func ListAllProducts(categoryId string) ([]Product, error) {
	productsTable := database.GetTables().ProductsTable
	ddbClient, err := database.NewDynamoDBClient(productsTable)
	if err != nil {
		return nil, err
	}
	data, err := ddbClient.Client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: &ddbClient.Table,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":categoryId": &types.AttributeValueMemberS{Value: categoryId},
		},
	})

	if err != nil {
		return nil, err
	}

	var products []Product
	err = attributevalue.UnmarshalListOfMaps(data.Items, &products)
	if err != nil {
		return nil, err
	}

	return products, nil
}

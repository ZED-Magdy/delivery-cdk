package models

import (
	"context"
	"fmt"

	"github.com/ZED-Magdy/delivery-cdk/lambda/database"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

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
		FilterExpression: aws.String("categoryId = :categoryId"),
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

func GetProductById(productId string) (*Product, error) {
	productsTable := database.GetTables().ProductsTable
	ddbClient, err := database.NewDynamoDBClient(productsTable)
	if err != nil {
		return nil, err
	}

	result, err := ddbClient.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &ddbClient.Table,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: productId},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, fmt.Errorf("product not found")
	}

	var product Product
	err = attributevalue.UnmarshalMap(result.Item, &product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

package models

import (
	"context"

	"github.com/ZED-Magdy/delivery-cdk/lambda/database"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Category represents a product category in the system
type Category struct {
	Id       string `json:"id" dynamodbav:"id"`
	Name     string `json:"name" dynamodbav:"name"`
	ImageUrl string `json:"imageUrl" dynamodbav:"imageUrl"`
}

// ListAllCategories retrieves all categories from the database
func ListAllCategories() ([]Category, error) {
	categoriesTable := database.GetTables().CategoriesTable
	ddbClient, err := database.NewDynamoDBClient(categoriesTable)
	if err != nil {
		return nil, err
	}

	data, err := ddbClient.Client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: &ddbClient.Table,
	})
	if err != nil {
		return nil, err
	}

	var categories []Category
	err = attributevalue.UnmarshalListOfMaps(data.Items, &categories)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

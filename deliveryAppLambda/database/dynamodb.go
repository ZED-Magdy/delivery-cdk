package database

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBClient struct {
	Client *dynamodb.Client
	Table  TableName
}

type Item struct {
	Name    string `json:"name" dynamodbav:"name"`
	Message string `json:"message" dynamodbav:"message"`
}
type TableName = string

type TableNames struct {
	AdsTable       TableName
	CategoriesTable TableName
	ProductsTable   TableName
	OrdersTable     TableName
	OrderItemsTable TableName
}

func GetTables() TableNames {
	return TableNames{
		AdsTable:      os.Getenv("ADS_TABLE"),
		CategoriesTable: os.Getenv("CATEGORIES_TABLE"),
		ProductsTable: os.Getenv("PRODUCTS_TABLE"),
		OrdersTable:   os.Getenv("ORDERS_TABLE"),
		OrderItemsTable: os.Getenv("ORDER_ITEMS_TABLE"),
	}
}

func NewDynamoDBClient(tableName TableName) (*DynamoDBClient, error) {
	
	cfg, err := config.LoadDefaultConfig(context.Background())
	if (err != nil) {
		return nil, err
	}

	return &DynamoDBClient{
		Client: dynamodb.NewFromConfig(cfg),
		Table:  tableName,
	}, nil
}
package database

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBClient struct {
	Client *dynamodb.Client
	Table  string
}

type Tables struct {
	AdsTable            string
	CategoriesTable     string
	ProductsTable       string
	OrdersTable         string
	OrderItemsTable     string
	DeliverAddressTable string
	UsersTable          string
}

func GetTables() Tables {
	return Tables{
		AdsTable:            os.Getenv("ADS_TABLE_NAME"),
		CategoriesTable:     os.Getenv("CATEGORIES_TABLE_NAME"),
		ProductsTable:       os.Getenv("PRODUCTS_TABLE_NAME"),
		OrdersTable:         os.Getenv("ORDERS_TABLE_NAME"),
		OrderItemsTable:     os.Getenv("ORDER_ITEMS_TABLE_NAME"),
		DeliverAddressTable: os.Getenv("DELIVER_ADDRESS_TABLE_NAME"),
		UsersTable:          os.Getenv("USERS_TABLE_NAME"),
	}
}

func NewDynamoDBClient(table string) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)
	return &DynamoDBClient{
		Client: client,
		Table:  table,
	}, nil
}
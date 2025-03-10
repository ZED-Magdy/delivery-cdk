package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DeliveryStackProps struct {
	awscdk.StackProps
}

func NewDeliveryStack(scope constructs.Construct, id string, props *DeliveryStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	AdsTable := awsdynamodb.NewTable(stack, jsii.String("Ads"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName: jsii.String("Ads"),
	})

	categoriesTable := awsdynamodb.NewTable(stack, jsii.String("categories"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName: jsii.String("categories"),
	})

	productsTable := awsdynamodb.NewTable(stack, jsii.String("products"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName: jsii.String("products"),
	})

	ordersTable := awsdynamodb.NewTable(stack, jsii.String("orders"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName: jsii.String("orders"),
	})

	orderItemsTable := awsdynamodb.NewTable(stack, jsii.String("orderItems"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName: jsii.String("orderItems"),
	})

	deliverAddressTable := awsdynamodb.NewTable(stack, jsii.String("deliverAddress"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName: jsii.String("deliverAddress"),
	})

	fn := awslambda.NewFunction(stack, jsii.String("deliveryApp"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Code: awslambda.Code_FromAsset(jsii.String("deliveryAppLambda/function.zip"), nil),
		Environment: &map[string]*string{
			"ADS_TABLE_NAME": AdsTable.TableName(),
			"CATEGORIES_TABLE_NAME": categoriesTable.TableName(),
			"PRODUCTS_TABLE_NAME": productsTable.TableName(),
			"ORDERS_TABLE_NAME": ordersTable.TableName(),
			"ORDER_ITEMS_TABLE_NAME": orderItemsTable.TableName(),
			"DELIVER_ADDRESS_TABLE_NAME": deliverAddressTable.TableName(),
		},
	})

	AdsTable.GrantReadData(fn)
	categoriesTable.GrantReadData(fn)
	productsTable.GrantReadData(fn)
	ordersTable.GrantReadWriteData(fn)
	orderItemsTable.GrantReadWriteData(fn)
	deliverAddressTable.GrantReadWriteData(fn)
	
	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewDeliveryStack(app, "DeliveryStack", &DeliveryStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return nil
}

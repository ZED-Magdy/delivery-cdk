package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func setupRoutes(api awsapigateway.LambdaRestApi) {
	api.Root().AddResource(jsii.String("ads"), nil).AddMethod(jsii.String("GET"), nil, nil)
	api.Root().AddResource(jsii.String("categories"), nil).AddMethod(jsii.String("GET"), nil, nil)
	api.Root().AddResource(jsii.String("products"), nil).AddResource(jsii.String("{categoryId}"), nil).AddMethod(jsii.String("GET"), nil, nil)
	
	users := api.Root().AddResource(jsii.String("users"), nil)
	users.AddResource(jsii.String("register"), nil).AddMethod(jsii.String("POST"), nil, nil)
	users.AddResource(jsii.String("send-otp"), nil).AddMethod(jsii.String("POST"), nil, nil)
	users.AddResource(jsii.String("verify-otp"), nil).AddMethod(jsii.String("POST"), nil, nil)
}

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

	usersTable := awsdynamodb.NewTable(stack, jsii.String("users"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName: jsii.String("users"),
	})

	usersTable.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String("PhoneIndex"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("phone"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		ProjectionType: awsdynamodb.ProjectionType_ALL,
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
			"USERS_TABLE_NAME": usersTable.TableName(),
			"JWT_SECRET": jsii.String("jwtsecret"), //FIXME: use aws secrets manager in production
		},
	})

	AdsTable.GrantReadData(fn)
	categoriesTable.GrantReadData(fn)
	productsTable.GrantReadData(fn)
	ordersTable.GrantReadWriteData(fn)
	orderItemsTable.GrantReadWriteData(fn)
	deliverAddressTable.GrantReadWriteData(fn)
	usersTable.GrantReadWriteData(fn)

	apiGateway := awsapigateway.NewLambdaRestApi(stack, jsii.String("deliveryAppApi"), &awsapigateway.LambdaRestApiProps{
		Handler: fn,
		Description: jsii.String("API Gateway for Delivery App Lambda"),
		DeployOptions: &awsapigateway.StageOptions{
			StageName: jsii.String("prod"),
		},
	})

	setupRoutes(apiGateway)
	
	awscdk.NewCfnOutput(stack, jsii.String("ApiEndpoint"), &awscdk.CfnOutputProps{
		Value:       apiGateway.Url(),
		Description: jsii.String("URL of the API Gateway"),
	})

	
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

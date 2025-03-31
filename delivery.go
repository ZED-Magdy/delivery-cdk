package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
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
	
	orders := api.Root().AddResource(jsii.String("orders"), nil)
	orders.AddMethod(jsii.String("POST"), nil, nil)
	orders.AddMethod(jsii.String("GET"), nil, nil)
	
	orderResource := orders.AddResource(jsii.String("{orderId}"), nil)
	orderResource.AddMethod(jsii.String("GET"), nil, nil)
	orderResource.AddResource(jsii.String("cancel"), nil).AddMethod(jsii.String("POST"), nil, nil)
	
	deliveryAddresses := api.Root().AddResource(jsii.String("delivery-addresses"), nil)
	deliveryAddresses.AddMethod(jsii.String("POST"), nil, nil)
	deliveryAddresses.AddMethod(jsii.String("GET"), nil, nil)
}

type DeliveryStackProps struct {
	awscdk.StackProps
}

// createDynamoTable creates a DynamoDB table with standard configuration
func createDynamoTable(stack awscdk.Stack, name string) awsdynamodb.Table {
	return awsdynamodb.NewTable(stack, jsii.String(name), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName: jsii.String(name),
	})
}

// grantLambdaTableAccess grants appropriate permissions to a Lambda function for a DynamoDB table
func grantLambdaTableAccess(table awsdynamodb.Table, lambdaFn awslambda.Function, readOnly bool) {
	if readOnly {
		table.GrantReadData(lambdaFn)
	} else {
		table.GrantReadWriteData(lambdaFn)
	}
}

func NewDeliveryStack(scope constructs.Construct, id string, props *DeliveryStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create DynamoDB tables
	tables := map[string]awsdynamodb.Table{
		"Ads":            createDynamoTable(stack, "Ads"),
		"Categories":     createDynamoTable(stack, "Categories"),
		"Products":       createDynamoTable(stack, "Products"),
		"Orders":         createDynamoTable(stack, "Orders"),
		"OrderItems":     createDynamoTable(stack, "OrderItems"),
		"DeliveryAddress": createDynamoTable(stack, "DeliveryAddress"),
		"Users":          createDynamoTable(stack, "Users"),
	}

	// Add GSI to Users table
	tables["Users"].AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String("PhoneIndex"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("phone"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		ProjectionType: awsdynamodb.ProjectionType_ALL,
	})

	// Create SQS queue and SNS topic
	ordersQueue := awssqs.NewQueue(stack, jsii.String("OrderQueue"), &awssqs.QueueProps{
		QueueName: jsii.String("OrderQueue"),
	})

	notificationTopic := awssns.NewTopic(stack, jsii.String("OrderStatusNotification"), &awssns.TopicProps{
		TopicName: jsii.String("OrderStatusNotification"),
	})

	// Prepare environment variables for Lambda functions
	baseEnvVars := map[string]*string{
		"ADS_TABLE_NAME":            tables["Ads"].TableName(),
		"CATEGORIES_TABLE_NAME":     tables["Categories"].TableName(),
		"PRODUCTS_TABLE_NAME":       tables["Products"].TableName(),
		"ORDERS_TABLE_NAME":         tables["Orders"].TableName(),
		"ORDER_ITEMS_TABLE_NAME":    tables["OrderItems"].TableName(),
		"DELIVERY_ADDRESS_TABLE_NAME": tables["DeliveryAddress"].TableName(),
		"USERS_TABLE_NAME":          tables["Users"].TableName(),
		"ORDER_QUEUE_URL":           ordersQueue.QueueUrl(),
		"ORDER_STATUS_NOTIFICATION_TOPIC_ARN": notificationTopic.TopicArn(),
	}

	// Main API Lambda function
	apiLambda := awslambda.NewFunction(stack, jsii.String("DeliveryApp"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Code:    awslambda.Code_FromAsset(jsii.String("deliveryAppLambda/function.zip"), nil),
		Environment: &map[string]*string{
			"ADS_TABLE_NAME":                    baseEnvVars["ADS_TABLE_NAME"],
			"CATEGORIES_TABLE_NAME":             baseEnvVars["CATEGORIES_TABLE_NAME"],
			"PRODUCTS_TABLE_NAME":               baseEnvVars["PRODUCTS_TABLE_NAME"],
			"ORDERS_TABLE_NAME":                 baseEnvVars["ORDERS_TABLE_NAME"],
			"ORDER_ITEMS_TABLE_NAME":            baseEnvVars["ORDER_ITEMS_TABLE_NAME"],
			"DELIVERY_ADDRESS_TABLE_NAME":       baseEnvVars["DELIVERY_ADDRESS_TABLE_NAME"],
			"USERS_TABLE_NAME":                  baseEnvVars["USERS_TABLE_NAME"],
			"ORDER_QUEUE_URL":                   baseEnvVars["ORDER_QUEUE_URL"],
			"ORDER_STATUS_NOTIFICATION_TOPIC_ARN": baseEnvVars["ORDER_STATUS_NOTIFICATION_TOPIC_ARN"],
			"JWT_SECRET":                        jsii.String("jwtsecret"), //FIXME: use aws secrets manager in production
		},
	})

	// Order processor Lambda function
	orderProcessorLambda := awslambda.NewFunction(stack, jsii.String("OrderProcessor"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("handlers.ProcessOrderQueue"),
		Code:    awslambda.Code_FromAsset(jsii.String("deliveryAppLambda/function.zip"), nil),
		Environment: &map[string]*string{
			"ORDERS_TABLE_NAME":                 baseEnvVars["ORDERS_TABLE_NAME"],
			"ORDER_ITEMS_TABLE_NAME":            baseEnvVars["ORDER_ITEMS_TABLE_NAME"],
			"DELIVERY_ADDRESS_TABLE_NAME":       baseEnvVars["DELIVERY_ADDRESS_TABLE_NAME"],
			"USERS_TABLE_NAME":                  baseEnvVars["USERS_TABLE_NAME"],
			"ORDER_QUEUE_URL":                   baseEnvVars["ORDER_QUEUE_URL"],
			"ORDER_STATUS_NOTIFICATION_TOPIC_ARN": baseEnvVars["ORDER_STATUS_NOTIFICATION_TOPIC_ARN"],
		},
	})

	// Grant permissions to API Lambda
	grantLambdaTableAccess(tables["Ads"], apiLambda, true) // Read-only
	grantLambdaTableAccess(tables["Categories"], apiLambda, true) // Read-only
	grantLambdaTableAccess(tables["Products"], apiLambda, true) // Read-only
	grantLambdaTableAccess(tables["Orders"], apiLambda, false) // Read-write
	grantLambdaTableAccess(tables["OrderItems"], apiLambda, false) // Read-write
	grantLambdaTableAccess(tables["DeliveryAddress"], apiLambda, false) // Read-write
	grantLambdaTableAccess(tables["Users"], apiLambda, false) // Read-write
	
	ordersQueue.GrantSendMessages(apiLambda)
	ordersQueue.GrantConsumeMessages(apiLambda)
	notificationTopic.GrantPublish(apiLambda)

	// Grant permissions to Order Processor Lambda
	grantLambdaTableAccess(tables["Orders"], orderProcessorLambda, false) // Read-write
	grantLambdaTableAccess(tables["OrderItems"], orderProcessorLambda, false) // Read-write
	grantLambdaTableAccess(tables["DeliveryAddress"], orderProcessorLambda, false) // Read-write
	grantLambdaTableAccess(tables["Users"], orderProcessorLambda, false) // Read-write
	
	ordersQueue.GrantConsumeMessages(orderProcessorLambda)
	notificationTopic.GrantPublish(orderProcessorLambda)

	// Create API Gateway
	apiGateway := awsapigateway.NewLambdaRestApi(stack, jsii.String("DeliveryAppApi"), &awsapigateway.LambdaRestApiProps{
		Handler: apiLambda,
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

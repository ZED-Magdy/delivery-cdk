# Delivery App

A serverless application for managing product deliveries built with AWS CDK and Go.

## Technology Stack

- **AWS Cloud Development Kit (CDK)** - Infrastructure as Code framework
- **AWS Lambda** - Serverless compute service for backend operations
- **Amazon DynamoDB** - NoSQL database for storing application data
- **Amazon SQS** - Queue service for order processing
- **Amazon SNS** - Notification service for order status updates
- **Go (Golang)** - Programming language for both infrastructure and business logic
- **AWS CloudFormation** - Provisioning and managing AWS resources

## Project Structure

```
delivery/
├── README.md              # Project documentation
├── cdk.json               # CDK configuration
├── delivery.go            # Main CDK infrastructure definition
├── deliveryAppLambda/     # Lambda function source code
│   ├── main.go            # Lambda handler implementation
│   └── function.zip       # Compiled Lambda function (generated)
```

## Database Structure

The application uses multiple DynamoDB tables to store different types of data:

| Table Name     | Partition Key | Description                         |
| -------------- | ------------- | ----------------------------------- |
| Ads            | id (string)   | Stores advertisement data           |
| categories     | id (string)   | Stores product categories           |
| products       | id (string)   | Stores product information          |
| orders         | id (string)   | Stores order details                |
| orderItems     | id (string)   | Stores items associated with orders |
| deliverAddress | id (string)   | Stores delivery addresses           |

## Setup and Deployment

### Prerequisites

1. Install the AWS CLI and configure your credentials
2. Install Go (1.18 or later recommended)
3. Install Node.js and npm (for CDK)
4. Install the CDK toolkit:
   ```
   npm install -g aws-cdk
   ```

### Building the Lambda Function

```bash
cd deliveryAppLambda
GOOS=linux GOARCH=amd64 go build -o  bootstrap
zip function.zip bootstrap
cd ..
```

### Deploying the Application

```bash
cdk synth
cdk bootstrap  # Only needed for the first time in an AWS account/region
cdk deploy
```

## Local Development

To test Lambda functions locally before deployment, you can use the AWS SAM CLI or create unit tests.

### Testing the Lambda Function

```bash
cd deliveryAppLambda
go test ./...
```

## Useful Commands

- `cdk deploy` Deploy this stack to your default AWS account/region
- `cdk diff` Compare deployed stack with current state
- `cdk synth` Emit the synthesized CloudFormation template
- `cdk destroy` Remove the deployed stack from your AWS account
- `go test` Run unit tests
- `go build` Build the Go application locally

## Accessing the Application

After deployment, the CDK will output the necessary information to access your application, such as API endpoints or other AWS resource identifiers.

## Monitoring and Logging

Lambda function logs are available in Amazon CloudWatch Logs. You can monitor application metrics in CloudWatch Dashboards.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

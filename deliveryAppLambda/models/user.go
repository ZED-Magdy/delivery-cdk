package models

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ZED-Magdy/delivery-cdk/lambda/database"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type User struct {
	ID          string    `json:"id" dynamodbav:"id"`
	Name        string    `json:"name" dynamodbav:"name"`
	Phone       string    `json:"phone" dynamodbav:"phone"`
	OTP         string    `json:"otp,omitempty" dynamodbav:"otp"`
	OTPExpiresAt time.Time `json:"otp_expires_at,omitempty" dynamodbav:"otp_expires_at"`
}

type UserRegistrationInput struct {
	Name  string `json:"name" validate:"required"`
	Phone string `json:"phone" validate:"required"`
}

type OTPVerificationInput struct {
	Phone string `json:"phone" validate:"required"`
	OTP   string `json:"otp" validate:"required"`
}

type SendOTPInput struct {
	Phone string `json:"phone" validate:"required"`
}

type Claims struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	jwt.RegisteredClaims
}

func RegisterUser(input UserRegistrationInput) (*User, error) {
	usersTable := database.GetTables().UsersTable
	ddbClient, err := database.NewDynamoDBClient(usersTable)
	if err != nil {
		return nil, err
	}

	keyEx := expression.Key("phone").Equal(expression.Value(input.Phone))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, err
	}

	response, err := ddbClient.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              &ddbClient.Table,
		IndexName:              aws.String("PhoneIndex"),
		KeyConditionExpression: expr.KeyCondition(),
		ExpressionAttributeNames: expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	if response.Count > 0 {
		return nil, errors.New("phone number already registered")
	}

	otp := "123456"

	user := &User{
		ID:          uuid.New().String(),
		Name:        input.Name,
		Phone:       input.Phone,
		OTP:         otp,
		OTPExpiresAt: time.Now().Add(2 * time.Minute),
	}

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = ddbClient.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &ddbClient.Table,
		Item:      item,
		ConditionExpression: aws.String("attribute_not_exists(phone)"),
	})
	
	if err != nil {
		var conditionalCheckFailedErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedErr) {
			return nil, errors.New("phone number already registered")
		}
		return nil, err
	}

	user.OTP = ""

	return user, nil
}

func VerifyOTP(input OTPVerificationInput) (*User, error) {
	usersTable := database.GetTables().UsersTable
	ddbClient, err := database.NewDynamoDBClient(usersTable)
	if err != nil {
		return nil, err
	}

	keyEx := expression.Key("phone").Equal(expression.Value(input.Phone))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, err
	}

	response, err := ddbClient.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              &ddbClient.Table,
		IndexName:              aws.String("PhoneIndex"),
		KeyConditionExpression: expr.KeyCondition(),
		ExpressionAttributeNames: expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	if response.Count == 0 {
		return nil, errors.New("user not found")
	}

	var users []User
	err = attributevalue.UnmarshalListOfMaps(response.Items, &users)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New("user not found")
	}

	user := users[0]

	if user.OTP != input.OTP {
		return nil, errors.New("invalid OTP")
	}

	if time.Now().After(user.OTPExpiresAt) {
		return nil, errors.New("OTP expired")
	}

	updateExpr := expression.Set(expression.Name("otp"), expression.Value(""))
	updateExpr = updateExpr.Set(expression.Name("otp_expires_at"), expression.Value(time.Time{}))
	
	expr, err = expression.NewBuilder().WithUpdate(updateExpr).Build()
	if err != nil {
		return nil, err
	}

	_, err = ddbClient.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: &ddbClient.Table,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: user.ID},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	user.OTP = ""
	user.OTPExpiresAt = time.Time{}

	return &user, nil
}

func SendOTP(input SendOTPInput) error {
	usersTable := database.GetTables().UsersTable
	ddbClient, err := database.NewDynamoDBClient(usersTable)
	if (err != nil) {
		return err
	}

	keyEx := expression.Key("phone").Equal(expression.Value(input.Phone))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if (err != nil) {
		return err
	}

	response, err := ddbClient.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              &ddbClient.Table,
		IndexName:              aws.String("PhoneIndex"),
		KeyConditionExpression: expr.KeyCondition(),
		ExpressionAttributeNames: expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if (err != nil) {
		return err
	}

	if (response.Count == 0) {
		return errors.New("user not found")
	}

	var users []User
	err = attributevalue.UnmarshalListOfMaps(response.Items, &users)
	if (err != nil) {
		return err
	}

	if (len(users) == 0) {
		return errors.New("user not found")
	}

	user := users[0]

	otp := "123456"
	otpExpiry := time.Now().Add(2 * time.Minute)

	updateExpr := expression.Set(expression.Name("otp"), expression.Value(otp))
	updateExpr = updateExpr.Set(expression.Name("otp_expires_at"), expression.Value(otpExpiry))
	
	expr, err = expression.NewBuilder().WithUpdate(updateExpr).Build()
	if (err != nil) {
		return err
	}

	_, err = ddbClient.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: &ddbClient.Table,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: user.ID},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	
	return err
}

func GetUserByID(userID string) (*User, error) {
	usersTable := database.GetTables().UsersTable
	ddbClient, err := database.NewDynamoDBClient(usersTable)
	if err != nil {
		return nil, err
	}

	response, err := ddbClient.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &ddbClient.Table,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(response.Item) == 0 {
		return nil, errors.New("user not found")
	}

	user := new(User)
	err = attributevalue.UnmarshalMap(response.Item, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserByPhone(phone string) (*User, error) {
	usersTable := database.GetTables().UsersTable
	ddbClient, err := database.NewDynamoDBClient(usersTable)
	if err != nil {
		return nil, err
	}

	keyEx := expression.Key("phone").Equal(expression.Value(phone))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, err
	}

	response, err := ddbClient.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              &ddbClient.Table,
		IndexName:              aws.String("PhoneIndex"),
		KeyConditionExpression: expr.KeyCondition(),
		ExpressionAttributeNames: expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	if response.Count == 0 {
		return nil, errors.New("user not found")
	}

	var users []User
	err = attributevalue.UnmarshalListOfMaps(response.Items, &users)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New("user not found")
	}

	return &users[0], nil
}

func GetAuthUser(request events.APIGatewayProxyRequest) (*User, error) {
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		return nil, errors.New("authorization header is required")
	}

	// Extract the token from the Authorization header
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("authorization header format must be Bearer {token}")
	}
	tokenString := parts[1]

	// Parse and validate the token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &User{
		ID:    claims.UserID,
		Name:  claims.Name,
		Phone: claims.Phone,
	}, nil
}
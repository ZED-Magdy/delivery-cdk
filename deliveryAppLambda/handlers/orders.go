package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/ZED-Magdy/delivery-cdk/lambda/models"
	"github.com/aws/aws-lambda-go/events"
)

type CreateOrderRequest struct {
	DeliveryAddressId string                 `json:"deliveryAddressId"`
	Items             []CreateOrderItemInput `json:"items"`
}

type CreateOrderItemInput struct {
	ProductId string  `json:"productId"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type OrderResponse struct {
	Order models.Order       `json:"order"`
	Items []models.OrderItem `json:"items,omitempty"`
}

// CreateOrder creates a new order with items
func CreateOrder(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Verify user authentication
	user, err := models.GetAuthUser(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "Unauthorized: " + err.Error(),
		}, nil
	}

	userId := user.ID

	// Parse request body
	var createReq CreateOrderRequest
	err = json.Unmarshal([]byte(request.Body), &createReq)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid request format: " + err.Error(),
		}, nil
	}

	// Calculate total from items
	var total float64
	for _, item := range createReq.Items {
		total += item.Price * float64(item.Quantity)
	}

	// Create the order
	order, err := models.CreateOrder(models.Order{
		UserId:            userId,
		Total:             total,
		Status:            models.StatusPending,
		DeliveryAddressId: createReq.DeliveryAddressId,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error creating order: " + err.Error(),
		}, nil
	}

	// Create order items
	orderItems := []models.OrderItem{}
	for _, item := range createReq.Items {
		orderItem, err := models.CreateOrderItem(models.OrderItem{
			OrderId:   order.Id,
			ProductId: item.ProductId,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
		})
		if err != nil {
			// If there's an error, we should ideally roll back the order
			// For simplicity, we're just returning an error
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       "Error creating order item: " + err.Error(),
			}, nil
		}
		orderItems = append(orderItems, *orderItem)
	}

	// Prepare response
	response := OrderResponse{
		Order: *order,
		Items: orderItems,
	}

	jsonBody, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error converting response to JSON",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Body:       string(jsonBody),
	}, nil
}

// CancelOrder cancels an order if it is in pending status
func CancelOrder(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Verify user authentication
	user, err := models.GetAuthUser(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "Unauthorized: " + err.Error(),
		}, nil
	}

	userId := user.ID

	// Get order ID from path parameters
	orderId := request.PathParameters["orderId"]
	if orderId == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Order ID is required",
		}, nil
	}

	// Get the order
	order, err := models.GetOrderById(orderId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "Order not found: " + err.Error(),
		}, nil
	}

	// Check if the order belongs to the authenticated user
	if order.UserId != userId {
		return events.APIGatewayProxyResponse{
			StatusCode: 403,
			Body:       "You can only cancel your own orders",
		}, nil
	}

	// Update order status to canceled
	updatedOrder, err := models.UpdateOrderStatus(orderId, models.StatusCanceled)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       err.Error(),
		}, nil
	}

	jsonBody, err := json.Marshal(updatedOrder)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error converting response to JSON",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonBody),
	}, nil
}

// GetUserOrders retrieves all orders for the authenticated user
func GetUserOrders(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Verify user authentication
	user, err := models.GetAuthUser(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "Unauthorized: " + err.Error(),
		}, nil
	}

	userId := user.ID

	// Get orders for the user
	orders, err := models.GetUserOrders(userId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Error retrieving orders: %v", err),
		}, nil
	}

	// For each order, get its items (optional, depending on your needs)
	var orderResponses []OrderResponse
	for _, order := range orders {
		orderCopy := order // Create a copy to avoid issues with pointers
		response := OrderResponse{
			Order: orderCopy,
		}
		orderResponses = append(orderResponses, response)
	}

	jsonBody, err := json.Marshal(orderResponses)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error converting response to JSON",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonBody),
	}, nil
}

// GetOrderDetails retrieves detailed information about a specific order
func GetOrderDetails(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Verify user authentication
	user, err := models.GetAuthUser(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "Unauthorized: " + err.Error(),
		}, nil
	}

	userId := user.ID

	// Get order ID from path parameters
	orderId := request.PathParameters["orderId"]
	if orderId == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Order ID is required",
		}, nil
	}

	// Get the order
	order, err := models.GetOrderById(orderId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "Order not found: " + err.Error(),
		}, nil
	}

	// Check if the order belongs to the authenticated user
	if order.UserId != userId {
		return events.APIGatewayProxyResponse{
			StatusCode: 403,
			Body:       "You can only view your own orders",
		}, nil
	}

	// Get order items
	items, err := models.GetOrderItems(orderId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error retrieving order items: " + err.Error(),
		}, nil
	}

	response := OrderResponse{
		Order: *order,
		Items: items,
	}

	jsonBody, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error converting response to JSON",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonBody),
	}, nil
}

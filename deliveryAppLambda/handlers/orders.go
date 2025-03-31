package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/ZED-Magdy/delivery-cdk/lambda/models"
	"github.com/ZED-Magdy/delivery-cdk/lambda/services"
	"github.com/aws/aws-lambda-go/events"
)

type CreateOrderRequest struct {
	DeliveryAddressId string              `json:"deliveryAddressId"`
	Items             []OrderItemRequest  `json:"items"`
}

type OrderItemRequest struct {
	ProductId string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type OrderResponse struct {
	Order models.Order       `json:"order"`
	Items []models.OrderItem `json:"items,omitempty"`
}

func CreateOrder(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user, err := models.GetAuthUser(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "Unauthorized: " + err.Error(),
		}, nil
	}

	userId := user.ID

	var createReq CreateOrderRequest
	err = json.Unmarshal([]byte(request.Body), &createReq)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid request format: " + err.Error(),
		}, nil
	}

	address, err := models.GetDeliveryAddressById(createReq.DeliveryAddressId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 422,
			Body:       "Invalid delivery address: " + err.Error(),
		}, nil
	}
	
	if address.UserId != userId {
		return events.APIGatewayProxyResponse{
			StatusCode: 403,
			Body:       "You can only use delivery addresses that belong to you",
		}, nil
	}

	var total float64
	var orderItems []models.OrderItem
	
	for _, itemReq := range createReq.Items {
		product, err := models.GetProductById(itemReq.ProductId)
		if (err != nil) {
			return events.APIGatewayProxyResponse{
				StatusCode: 422,
				Body:       fmt.Sprintf("Invalid product ID %s: %s", itemReq.ProductId, err.Error()),
			}, nil
		}
		
		itemTotal := product.Price * float64(itemReq.Quantity)
		total += itemTotal
		
		orderItems = append(orderItems, models.OrderItem{
			ProductId: product.Id,
			Name:      product.Name,
			Price:     product.Price,
			Quantity:  itemReq.Quantity,
		})
	}

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

	var savedOrderItems []models.OrderItem
	for _, item := range orderItems {
		item.OrderId = order.Id
		savedItem, err := models.CreateOrderItem(item)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       "Error creating order item: " + err.Error(),
			}, nil
		}
		savedOrderItems = append(savedOrderItems, *savedItem)
	}

	// After the order is created successfully
	// Add this after the order and order items are saved successfully
	err = services.SendOrderToQueue(order.Id, string(models.StatusPending), userId)
	if err != nil {
		// Log the error but don't fail the order creation
		fmt.Printf("Error sending order to queue: %v\n", err)
	}

	response := OrderResponse{
		Order: *order,
		Items: savedOrderItems,
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

func CancelOrder(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user, err := models.GetAuthUser(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "Unauthorized: " + err.Error(),
		}, nil
	}

	userId := user.ID

	orderId := request.PathParameters["orderId"]
	if orderId == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Order ID is required",
		}, nil
	}

	order, err := models.GetOrderById(orderId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "Order not found: " + err.Error(),
		}, nil
	}

	if order.UserId != userId {
		return events.APIGatewayProxyResponse{
			StatusCode: 403,
			Body:       "You can only cancel your own orders",
		}, nil
	}

	updatedOrder, err := models.UpdateOrderStatus(orderId, models.StatusCanceled)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       err.Error(),
		}, nil
	}

	// Send the updated order status to the queue for processing
	err = services.SendOrderToQueue(orderId, string(models.StatusCanceled), userId)
	if err != nil {
		// Log the error but don't fail the cancel operation
		fmt.Printf("Error sending canceled order to queue: %v\n", err)
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

func GetUserOrders(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user, err := models.GetAuthUser(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "Unauthorized: " + err.Error(),
		}, nil
	}

	userId := user.ID

	orders, err := models.GetUserOrders(userId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Error retrieving orders: %v", err),
		}, nil
	}

	var orderResponses []OrderResponse
	for _, order := range orders {
		orderCopy := order
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

func GetOrderDetails(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	user, err := models.GetAuthUser(request)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "Unauthorized: " + err.Error(),
		}, nil
	}

	userId := user.ID

	orderId := request.PathParameters["orderId"]
	if orderId == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Order ID is required",
		}, nil
	}

	order, err := models.GetOrderById(orderId)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       "Order not found: " + err.Error(),
		}, nil
	}

	if order.UserId != userId {
		return events.APIGatewayProxyResponse{
			StatusCode: 403,
			Body:       "You can only view your own orders",
		}, nil
	}

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

package router

import (
	"github.com/aws/aws-lambda-go/events"
)

type RouteHandler func(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type MiddlewareFunc func(RouteHandler) RouteHandler

type Route struct {
	Path       string
	Method     string
	Handler    RouteHandler
	IsResource bool
	Middleware []MiddlewareFunc
}

type Router struct {
	routes []Route
	globalMiddleware []MiddlewareFunc
}

func NewRouter() *Router {
	return &Router{
		routes: []Route{},
		globalMiddleware: []MiddlewareFunc{},
	}
}

func (r *Router) Use(middleware ...MiddlewareFunc) {
	r.globalMiddleware = append(r.globalMiddleware, middleware...)
}

func (r *Router) Add(path, method string, handler RouteHandler, middleware ...MiddlewareFunc) {
	r.routes = append(r.routes, Route{
		Path:       path,
		Method:     method,
		Handler:    handler,
		Middleware: middleware,
	})
}

func applyMiddleware(handler RouteHandler, middleware []MiddlewareFunc) RouteHandler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

func (r *Router) Match(request events.APIGatewayProxyRequest) (RouteHandler, bool) {
	for _, route := range r.routes {
		matches := request.Resource == route.Path && (route.Method == "" || request.HTTPMethod == route.Method)
		
		if matches {
			allMiddleware := append([]MiddlewareFunc{}, r.globalMiddleware...)
			allMiddleware = append(allMiddleware, route.Middleware...)
			
			finalHandler := applyMiddleware(route.Handler, allMiddleware)
			return finalHandler, true
		}
	}
	return nil, false
}

func AdaptMiddleware(middleware func(handlerFunc func(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)) func(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)) MiddlewareFunc {
	return func(next RouteHandler) RouteHandler {
		return middleware(next)
	}
}

func AdaptAuthMiddleware(middleware interface{}) MiddlewareFunc {
	return func(next RouteHandler) RouteHandler {
		return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			authNext := func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
				return next(req)
			}
			
			if authMiddleware, ok := middleware.(func(func(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)) func(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)); ok {
				handler := authMiddleware(authNext)
				return handler(request)
			}
			
			return next(request)
		}
	}
}

package middlewares

import (
	"github.com/stayfatal/VK-pub-sub/pkg/logger"
	"google.golang.org/grpc"
)

type (
	// UnaryMiddleware works with classic rpc
	UnaryMiddleware func(logger *logger.Logger, handler grpc.UnaryHandler) grpc.UnaryHandler

	// StreamMiddleware works with streams
	StreamMiddleware func(logger *logger.Logger, handler grpc.StreamHandler) grpc.StreamHandler
)

// left middleware wraps right middleware
func BuildUnaryChain(logger *logger.Logger, handler grpc.UnaryHandler, middlewares ...UnaryMiddleware) grpc.UnaryHandler {
	var wrapped grpc.UnaryHandler = middlewares[len(middlewares)-1](logger, handler)
	for i := len(middlewares) - 2; i >= 0; i-- {
		wrapped = middlewares[i](logger, wrapped)
	}
	return wrapped
}

// left middleware wraps right middleware
func BuildStreamChain(logger *logger.Logger, handler grpc.StreamHandler, middlewares ...StreamMiddleware) grpc.StreamHandler {
	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](logger, wrapped)
	}
	return wrapped
}

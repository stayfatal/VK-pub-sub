package middlewares

import (
	"context"

	"github.com/stayfatal/VK-pub-sub/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryRecoverer(logger *logger.Logger, handler grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req any) (any, error) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(nil, zap.Any("error", err))
			}
		}()

		return handler(ctx, req)
	}
}

func StreamRecoverer(logger *logger.Logger, handler grpc.StreamHandler) grpc.StreamHandler {
	return func(srv interface{}, stream grpc.ServerStream) error {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(nil, zap.Any("error", err))
			}
		}()
		return handler(srv, stream)
	}
}

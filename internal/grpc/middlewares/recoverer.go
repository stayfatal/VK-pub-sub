package middlewares

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryRecoverer(logger *zap.Logger, handler grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req any) (any, error) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("recovered", zap.Any("error", err))
			}
		}()

		return handler(ctx, req)
	}
}

func StreamRecoverer(logger *zap.Logger, handler grpc.StreamHandler) grpc.StreamHandler {
	return func(srv interface{}, stream grpc.ServerStream) error {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("stream panic recovered", zap.Any("error", err))
			}
		}()
		return handler(srv, stream)
	}
}

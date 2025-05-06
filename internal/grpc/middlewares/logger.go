package middlewares

import (
	"context"
	"time"

	"github.com/stayfatal/VK-pub-sub/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryLogger(logger *logger.Logger, handler grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req any) (any, error) {
		startTime := time.Now()
		resp, err := handler(ctx, req)
		method, _ := grpc.Method(ctx)

		logger.Info(
			"Processed request",
			zap.String("duration", time.Since(startTime).String()),
			zap.String("method", method),
			zap.String("code", status.Convert(err).Code().String()),
		)
		return resp, err
	}
}

func StreamLogger(logger *logger.Logger, handler grpc.StreamHandler) grpc.StreamHandler {
	return func(srv interface{}, stream grpc.ServerStream) error {
		startTime := time.Now()
		method, _ := grpc.MethodFromServerStream(stream)

		logger.Info(
			"Stream started",
			zap.String("method", method),
		)

		err := handler(srv, stream)

		logger.Info(
			"Stream finished",
			zap.String("duration", time.Since(startTime).String()),
			zap.String("method", method),
			zap.String("code", status.Convert(err).Code().String()),
		)

		return err
	}
}

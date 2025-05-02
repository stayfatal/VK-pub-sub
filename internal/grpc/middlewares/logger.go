package middlewares

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryLogger(logger *zap.Logger, handler grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req any) (any, error) {
		startTime := time.Now()
		resp, err := handler(ctx, req)
		method, _ := grpc.Method(ctx)

		fields := []zap.Field{
			zap.String("duration", time.Since(startTime).String()),
			zap.String("method", method),
		}

		if err != nil {
			fields = append(fields, zap.String("error", err.Error()))
		}

		logger.Log(
			getLevel(err),
			"Processed request",
			fields...,
		)
		return resp, err
	}
}

func StreamLogger(logger *zap.Logger, handler grpc.StreamHandler) grpc.StreamHandler {
	return func(srv interface{}, stream grpc.ServerStream) error {
		startTime := time.Now()
		method, _ := grpc.MethodFromServerStream(stream)

		logger.Info(
			"Stream started",
			zap.String("method", method),
		)

		err := handler(srv, stream)

		fields := []zap.Field{
			zap.String("duration", time.Since(startTime).String()),
			zap.String("method", method),
		}

		if err != nil {
			fields = append(fields, zap.String("error", err.Error()))
		}

		logger.Log(
			getLevel(err),
			"Stream finished",
			fields...,
		)

		return err
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/stayfatal/VK-pub-sub/gen/pubsubpb"
	server "github.com/stayfatal/VK-pub-sub/internal/grpc"
	"github.com/stayfatal/VK-pub-sub/internal/service"
	"github.com/stayfatal/VK-pub-sub/pkg/logger"
	"github.com/stayfatal/VK-pub-sub/pkg/pubsub"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("cant load env file")
	}

	port := os.Getenv("PORT")

	logLevel, err := zap.ParseAtomicLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Println("cant parse log level variable in config setting to info level")
	}

	logger, err := logger.NewLogger(logLevel)
	if err != nil {
		log.Fatal(err)
	}

	pubsub := pubsub.NewPubSub()
	defer pubsub.Close(context.Background())

	svc := service.NewService(logger, pubsub)

	srv := server.NewPubSubServer(svc, logger)

	grpcSrv := grpc.NewServer()
	defer grpcSrv.GracefulStop()

	pubsubpb.RegisterPubSubServer(grpcSrv, srv)

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Fatal("error while starting server", zap.String("error", err.Error()))
	}

	if err := grpcSrv.Serve(l); err != nil {
		logger.Fatal("error while serving grpc", zap.String("error", err.Error()))
	}
}

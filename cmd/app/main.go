package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/stayfatal/VK-pub-sub/config"
	"github.com/stayfatal/VK-pub-sub/gen/pubsubpb"
	server "github.com/stayfatal/VK-pub-sub/internal/grpc"
	"github.com/stayfatal/VK-pub-sub/internal/service"
	"github.com/stayfatal/VK-pub-sub/pkg/logger"
	"github.com/stayfatal/VK-pub-sub/pkg/pubsub"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	config, err := config.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	logger, err := logger.NewLogger(config.Logger.Level)
	if err != nil {
		log.Fatal(err)
		return
	}

	pubsub := pubsub.NewPubSub()
	defer pubsub.Close(context.Background())

	svc := service.NewService(pubsub)

	srv := server.NewPubSubServer(svc, logger)

	grpcSrv := grpc.NewServer()
	defer grpcSrv.GracefulStop()

	pubsubpb.RegisterPubSubServer(grpcSrv, srv)

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Server.Port))
	if err != nil {
		logger.Error(err)
		return
	}
	defer l.Close()

	srvErr := make(chan error)
	go func() {
		logger.Info("server is listening", zap.String("port", config.Server.Port))
		if err := grpcSrv.Serve(l); err != nil {
			srvErr <- err
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	select {
	case sig := <-c:
		logger.Info("Server stopped", zap.String("signal", sig.String()))
	case <-srvErr:
		logger.Error(err)
	}
}

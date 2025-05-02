package grpc

import (
	"context"

	"github.com/stayfatal/VK-pub-sub/gen/pubsubpb"
	"github.com/stayfatal/VK-pub-sub/internal/grpc/middlewares"
	"github.com/stayfatal/VK-pub-sub/pkg/pubsub"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service interface {
	Publish(context.Context, string, interface{}) error
	Subscribe(string) (chan interface{}, pubsub.Subscription, error)
}

type Server struct {
	pubsubpb.UnimplementedPubSubServer

	PublishHandler   grpc.UnaryHandler
	SubscribeHandler grpc.StreamHandler
}

func NewPubSubServer(svc Service, logger *zap.Logger) *Server {
	handlersManager := Handlers{
		svc:    svc,
		logger: logger,
	}

	return &Server{
		PublishHandler: middlewares.BuildUnaryChain(
			logger,
			handlersManager.publishHandler,
			middlewares.UnaryRecoverer, middlewares.UnaryLogger,
		),
		SubscribeHandler: middlewares.BuildStreamChain(
			logger,
			handlersManager.subscribeHandler,
			middlewares.StreamRecoverer, middlewares.StreamLogger,
		),
	}
}

func (s *Server) Publish(ctx context.Context, req *pubsubpb.PublishRequest) (*emptypb.Empty, error) {
	_, err := s.PublishHandler(ctx, req)
	return &emptypb.Empty{}, err
}

func (s *Server) Subscribe(req *pubsubpb.SubscribeRequest, stream pubsubpb.PubSub_SubscribeServer) error {
	return s.SubscribeHandler(req, stream)
}

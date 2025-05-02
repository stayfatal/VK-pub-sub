package grpc

import (
	"context"

	"github.com/stayfatal/VK-pub-sub/domain"
	"github.com/stayfatal/VK-pub-sub/gen/pubsubpb"
	"github.com/stayfatal/VK-pub-sub/pkg/myerrors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Handlers struct {
	svc    Service
	logger *zap.Logger
}

func (h *Handlers) publishHandler(ctx context.Context, req any) (any, error) {
	request, ok := req.(*pubsubpb.PublishRequest)
	if !ok {
		return nil, myerrors.New(domain.AssertionErr, zap.ErrorLevel)
	}
	return nil, h.svc.Publish(ctx, request.Key, request.Data)
}

func (h *Handlers) subscribeHandler(req any, st grpc.ServerStream) error {
	request, ok := req.(*pubsubpb.SubscribeRequest)
	if !ok {
		return myerrors.New(domain.AssertionErr, zap.ErrorLevel)
	}
	ch, sub, err := h.svc.Subscribe(request.Key)
	if err != nil {
		return err
	}

	stream, ok := st.(pubsubpb.PubSub_SubscribeServer)
	if !ok {
		return myerrors.New(domain.AssertionErr, zap.ErrorLevel)
	}

	for {
		select {
		case <-stream.Context().Done():
			sub.Unsubscribe()
			return myerrors.New(stream.Context().Err(), zap.InfoLevel)
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			data, ok := msg.(string)
			if !ok {
				h.logger.Error("error while handling message", zap.String("error", domain.AssertionErr.Error()))
				continue
			}

			err := stream.Send(&pubsubpb.Event{
				Data: data,
			})
			if err != nil {
				h.logger.Error("error while sending message", zap.String("error", err.Error()))
				continue
			}
		}

	}
}

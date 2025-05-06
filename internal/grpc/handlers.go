package grpc

import (
	"context"

	"github.com/stayfatal/VK-pub-sub/domain"
	"github.com/stayfatal/VK-pub-sub/gen/pubsubpb"
	"github.com/stayfatal/VK-pub-sub/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handlers struct {
	svc    Service
	logger *logger.Logger
}

func (h *Handlers) publishHandler(ctx context.Context, req any) (any, error) {
	request, ok := req.(*pubsubpb.PublishRequest)
	if !ok {
		h.logger.Error(domain.AssertionErr)
		return nil, status.Error(codes.Internal, domain.AssertionErr.Error())
	}

	err := h.svc.Publish(ctx, request.Key, request.Data)
	if err != nil {
		h.logger.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}

func (h *Handlers) subscribeHandler(req any, st grpc.ServerStream) error {
	request, ok := req.(*pubsubpb.SubscribeRequest)
	if !ok {
		h.logger.Error(domain.AssertionErr)
		return status.Error(codes.Internal, domain.AssertionErr.Error())
	}
	ch, sub, err := h.svc.Subscribe(request.Key)
	if err != nil {
		h.logger.Error(err)
		return status.Error(codes.Internal, err.Error())
	}

	stream, ok := st.(pubsubpb.PubSub_SubscribeServer)
	if !ok {
		h.logger.Error(domain.AssertionErr)
		return status.Error(codes.Internal, domain.AssertionErr.Error())
	}

	for {
		select {
		case <-stream.Context().Done():
			sub.Unsubscribe()
			return status.Error(codes.Canceled, stream.Context().Err().Error())
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			data, ok := msg.(string)
			if !ok {
				h.logger.Error(domain.AssertionErr)
				continue
			}

			err := stream.Send(&pubsubpb.Event{
				Data: data,
			})
			if err != nil {
				h.logger.Error(err)
				continue
			}
		}

	}
}

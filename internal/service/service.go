package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stayfatal/VK-pub-sub/pkg/myerrors"
	"github.com/stayfatal/VK-pub-sub/pkg/pubsub"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
	pubsub pubsub.PubSub
}

func NewService(logger *zap.Logger, pubsub pubsub.PubSub) *Service {
	return &Service{logger: logger, pubsub: pubsub}
}

func (s *Service) Publish(_ context.Context, subject string, msg interface{}) error {
	err := s.pubsub.Publish(subject, msg)
	// info level cuz Publish return error only if user input is wrong
	return errors.Wrap(myerrors.New(err, zap.InfoLevel), "error while publishing")
}

func (s *Service) Subscribe(subject string) (chan interface{}, pubsub.Subscription, error) {
	ch := make(chan interface{}, 1)
	sub, err := s.pubsub.Subscribe(subject, func(msg interface{}) {
		ch <- msg
	})
	// info level cuz Subscribe return error only if user input is wrong
	return ch, sub, errors.Wrap(myerrors.New(err, zap.InfoLevel), "error while subscribing to subject")
}

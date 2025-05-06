package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stayfatal/VK-pub-sub/pkg/pubsub"
)

type Service struct {
	pubsub pubsub.PubSub
}

func NewService(pubsub pubsub.PubSub) *Service {
	return &Service{pubsub: pubsub}
}

func (s *Service) Publish(_ context.Context, subject string, msg interface{}) error {
	return errors.Wrap(s.pubsub.Publish(subject, msg), "error while publishing")
}

func (s *Service) Subscribe(subject string) (chan interface{}, pubsub.Subscription, error) {
	ch := make(chan interface{}, 1)
	sub, err := s.pubsub.Subscribe(subject, func(msg interface{}) {
		ch <- msg
	})
	return ch, sub, errors.Wrap(err, "error while subscribing to subject")
}

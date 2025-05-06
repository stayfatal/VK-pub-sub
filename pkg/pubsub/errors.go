package pubsub

import "github.com/pkg/errors"

var (
	ErrClosed = errors.New("closed was called")
)

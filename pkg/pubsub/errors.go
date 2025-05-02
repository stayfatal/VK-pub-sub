package pubsub

import "github.com/pkg/errors"

var (
	ErrNonExistentSubject = errors.New("topics didn't contain such subject")
)

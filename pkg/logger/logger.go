package logger

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func NewLogger(level zap.AtomicLevel) (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = level
	logger, err := cfg.Build(zap.AddStacktrace(zap.ErrorLevel))
	return logger, errors.Wrap(err, "error while creating logger")
}

func NewMock() (*zap.Logger, error) {
	return zap.New(nil), nil
}

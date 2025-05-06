package logger

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Logger struct {
	*zap.Logger
}

func NewLogger(level zap.AtomicLevel, opts ...zap.Option) (*Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = level
	logger, err := cfg.Build(opts...)
	return &Logger{Logger: logger}, errors.Wrap(err, "error while creating logger")
}

func NewMock() (*zap.Logger, error) {
	return zap.New(nil), nil
}

func (l *Logger) Error(err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.String("error", err.Error()))
	}
	l.Logger.Error("error occurred", fields...)
}

func MockInit() *Logger {
	return &Logger{zap.NewNop()}
}

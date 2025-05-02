package middlewares

import (
	"github.com/pkg/errors"
	"github.com/stayfatal/VK-pub-sub/pkg/myerrors"
	"go.uber.org/zap/zapcore"
)

func getLevel(err error) zapcore.Level {
	if err == nil {
		return zapcore.InfoLevel
	}

	myErr, ok := errors.Unwrap(err).(*myerrors.Err)
	if !ok {
		return zapcore.InfoLevel
	}
	return myErr.Level
}

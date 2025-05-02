package myerrors

// this package provides better error handling by adding log level to error
// so you don't need to assert error and then log

import "go.uber.org/zap/zapcore"

type Err struct {
	err   error
	Level zapcore.Level
}

func New(err error, level zapcore.Level) error {
	if err == nil {
		return nil
	}
	return &Err{err: err, Level: level}
}

func (e *Err) Error() string {
	return e.err.Error()
}

func (e *Err) Unwrap() error {
	return e.err
}

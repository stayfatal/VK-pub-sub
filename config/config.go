package config

import (
	"os"

	"github.com/subosito/gotenv"
	"go.uber.org/zap"
)

type Config struct {
	Server ServerConfig
	Logger LoggerConfig
}

type ServerConfig struct {
	Port string
}

type LoggerConfig struct {
	Level zap.AtomicLevel
}

func Load(path string) (*Config, error) {
	err := gotenv.Load(path)
	if err != nil {
		return nil, err
	}

	port := "SERVER_PORT"
	logLevel := "LOG_LEVEL"

	level, err := zap.ParseAtomicLevel(os.Getenv(logLevel))
	if err != nil {
		return nil, err
	}

	return &Config{
		Server: ServerConfig{
			Port: os.Getenv(port),
		},
		Logger: LoggerConfig{
			Level: level,
		},
	}, nil
}

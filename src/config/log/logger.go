package log

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger     *zap.Logger
	onceLogger sync.Once
)

func GetLogger() *zap.Logger {
	onceLogger.Do(func() {
		var err error

		config := zap.NewDevelopmentConfig()

		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.MessageKey = "message"
		config.EncoderConfig.CallerKey = ""
		config.EncoderConfig.NameKey = ""

		logger, err = config.Build()
		if err != nil {
			panic(err)
		}
	})
	return logger
}

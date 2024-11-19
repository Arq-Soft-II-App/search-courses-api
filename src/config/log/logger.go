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

func getLogLevel() zapcore.Level {
	// Puedes hacer esto configurable mediante variables de entorno
	return zapcore.DebugLevel
}

func GetLogger() *zap.Logger {
	onceLogger.Do(func() {
		config := zap.NewProductionConfig()

		// Configuración personalizada
		config.Level = zap.NewAtomicLevelAt(getLogLevel())
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.LevelKey = "level"
		config.EncoderConfig.MessageKey = "message"
		config.EncoderConfig.CallerKey = ""
		config.EncoderConfig.StacktraceKey = ""
		config.EncoderConfig.NameKey = ""

		// Formato del mensaje
		config.OutputPaths = []string{"stdout"}
		config.Encoding = "console"

		var err error
		logger, err = config.Build(
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
		if err != nil {
			panic(err)
		}
	})
	return logger
}

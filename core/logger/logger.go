package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Get returns the initialized logger instance
func Get() (logger *zap.Logger) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.CallerKey = "caller"
	config.DisableStacktrace = true

	// Will not give error as all the configs are defined as constants
	logger, _ = config.Build()

	return logger
}

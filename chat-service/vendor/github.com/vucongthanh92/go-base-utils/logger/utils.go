package logger

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

func getZapLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case DebugLevel:
		return zapcore.DebugLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	case PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

func getLog(logs []Logger) Logger {
	var logger Logger
	if len(logs) > 0 {
		logger = logs[0]
	} else {
		logger = defaultLogger
	}
	return logger
}

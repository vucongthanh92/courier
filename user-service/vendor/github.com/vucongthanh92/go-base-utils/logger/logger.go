package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	//DebugLevel has verbose message
	DebugLevel = "debug"
	//InfoLevel is default log level
	InfoLevel = "info"
	//WarnLevel is for log messages about possible issues
	WarnLevel = "warn"
	//ErrorLevel is for log errors
	ErrorLevel = "error"
	//FatalLevel is for log fatal messages. The sytem shutsdown after log the message.
	FatalLevel = "fatal"
	//PanicLevel logs a message, then panics.
	PanicLevel = "panic"
)

// Logger is the fundamental interface for all log operations. Log creates a
// log event from keyvals, a variadic sequence of alternating keys and values.
// Implementations must be safe for concurrent use by multiple goroutines. In
// particular, any implementation of Logger that appends to keyvals or
// modifies or retains any of its elements must make a copy first.
type Logger interface {
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)
	Panic(msg string, fields ...zapcore.Field)
	Log(keyvals ...interface{}) error
	With(fields ...zapcore.Field) Logger
	WithOptions(opts ...zap.Option) Logger
	GetZapLogger() *zap.Logger
}

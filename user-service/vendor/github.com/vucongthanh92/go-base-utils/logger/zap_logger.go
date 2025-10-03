package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(consoleLevel string) Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.LevelKey = "lvl"
	encoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder // zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	if consoleLevel == "" ||
		(consoleLevel != DebugLevel && consoleLevel != InfoLevel && consoleLevel != WarnLevel && consoleLevel != ErrorLevel) {
		consoleLevel = InfoLevel
	}

	logger := zap.New(
		zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), getZapLevel(consoleLevel)),
		zap.AddCallerSkip(-1),
		zap.AddCaller())

	logger = logger.WithOptions(zap.AddCallerSkip(2))
	return zapLogger{logger}
}

// Debug logs an debug msg with fields
func (l zapLogger) Debug(msg string, fields ...zapcore.Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs an info msg with fields
func (l zapLogger) Info(msg string, fields ...zapcore.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs an warn msg with fields
func (l zapLogger) Warn(msg string, fields ...zapcore.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error msg with fields
func (l zapLogger) Error(msg string, fields ...zapcore.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs a fatal error msg with fields
func (l zapLogger) Fatal(msg string, fields ...zapcore.Field) {
	l.logger.Fatal(msg, fields...)
}

// Panic logs an panic msg with fields
func (l zapLogger) Panic(msg string, fields ...zapcore.Field) {
	l.logger.Panic(msg, fields...)
}

func (l zapLogger) Log(keyvals ...interface{}) error {
	l.logger.WithOptions(zap.AddCallerSkip(2)).Sugar().Infow("log", keyvals...)
	return nil
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (l zapLogger) With(fields ...zapcore.Field) Logger {
	return zapLogger{logger: l.logger.With(fields...)}
}

func (l zapLogger) WithOptions(opts ...zap.Option) Logger {
	return zapLogger{logger: l.logger.WithOptions(opts...)}
}

func (l zapLogger) GetZapLogger() *zap.Logger {
	return l.logger
}

package logger

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type traceLogger struct {
	l otelzap.LoggerWithCtx
}

func (t traceLogger) Debug(msg string, fields ...zapcore.Field) {
	t.l.Debug(msg, fields...)
}

func (t traceLogger) Info(msg string, fields ...zapcore.Field) {
	t.l.Info(msg, fields...)
}

func (t traceLogger) Warn(msg string, fields ...zapcore.Field) {
	t.l.Warn(msg, fields...)
}

func (t traceLogger) Error(msg string, fields ...zapcore.Field) {
	t.l.Error(msg, fields...)
}

func (t traceLogger) Fatal(msg string, fields ...zapcore.Field) {
	t.l.Fatal(msg, fields...)
}

func (t traceLogger) Panic(msg string, fields ...zapcore.Field) {
	t.l.Panic(msg, fields...)
}

func (t traceLogger) Log(keyvals ...interface{}) error {
	t.l.WithOptions(zap.AddCallerSkip(2)).Sugar().Infow("log", keyvals...)
	return nil
}

func (t traceLogger) With(fields ...zapcore.Field) Logger {
	t.l.ZapLogger().With(fields...)
	return t
}

func (t traceLogger) WithOptions(opts ...zap.Option) Logger {
	t.l = t.l.WithOptions(opts...)
	return t
}

func (t traceLogger) GetZapLogger() *zap.Logger {
	return t.l.ZapLogger()
}

// WithTrace  use logger with tracing context
func WithTrace(ctx context.Context, logg ...Logger) Logger {
	logger := getLog(logg).GetZapLogger()
	log := otelzap.New(logger.With(zap.String("request_id", GetTraceIDFromContext(ctx))),
		otelzap.WithMinLevel(zapcore.InfoLevel),
		otelzap.WithStackTrace(true),
		otelzap.WithCaller(true),
		otelzap.WithCallerDepth(2),
		otelzap.WithErrorStatusLevel(zapcore.ErrorLevel)).Ctx(ctx)
	return traceLogger{log}
}

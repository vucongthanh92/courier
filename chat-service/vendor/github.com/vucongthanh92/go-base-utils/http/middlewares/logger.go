package middlewares

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"

	"github.com/vucongthanh92/go-base-utils/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Logging(skip ...string) gin.HandlerFunc {
	skip = append(skip, "/metrics", "/auth/token")
	return func(c *gin.Context) {
		for _, s := range skip {
			if s == c.Request.URL.Path {
				c.Next()
				return
			}
		}

		start := time.Now()
		var body []byte
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ = io.ReadAll(tee)
		c.Request.Body = io.NopCloser(&buf)
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		c.Next()
		fields := []zapcore.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("response", blw.body.String()),
		}

		if len(c.Request.URL.RawQuery) > 0 {
			fields = append(fields, zap.String("query", c.Request.URL.RawQuery))
		}

		if requestID := c.Writer.Header().Get("X-Request-Id"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		if span := trace.SpanFromContext(c.Request.Context()).SpanContext(); span.IsValid() {
			fields = append(fields, zap.String("trace_id", span.TraceID().String()))
		}

		if len(body) > 0 {
			fields = append(fields, zap.String("request", string(body)))
		}

		fields = append(fields, zap.Duration("latency", time.Since(start)))
		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			for _, e := range c.Errors.Errors() {
				logger.Error(e, fields...)
			}
		} else {
			logger.Info(c.Request.URL.Path, fields...)
		}
	}
}

func Recovery(logger logger.Logger) gin.HandlerFunc {
	return ginzap.RecoveryWithZap(logger.GetZapLogger(), true)
}

func Tracing(name string) gin.HandlerFunc {
	return otelgin.Middleware(name, otelgin.WithPropagators(otel.GetTextMapPropagator()))
}

func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the traceID from the response header
		traceID := c.Writer.Header().Get("X-Request-Id")

		// Create a new context with the traceID
		ctx := context.WithValue(c.Request.Context(), logger.TraceKey, traceID)

		// Set the context back to the request
		c.Request = c.Request.WithContext(ctx)

		// Continue to the next middleware/handler
		c.Next()
	}
}

func RequestId() gin.HandlerFunc {
	return requestid.New()
}

func Gzip() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression)
}

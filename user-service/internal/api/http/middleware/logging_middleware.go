package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vucongthanh92/courier/user-service/internal/repository/external/loki"
	"go.opentelemetry.io/otel/trace"
)

var sensitiveHeaders = map[string]struct{}{
	"Authorization": {}, "Cookie": {}, "Set-Cookie": {},
}

var sensitiveKeys = []string{"password", "token", "otp", "refresh_token", "pass"}

func RequestLoggingMiddleware(lc loki.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Set("request_id", reqID)

		// copy body (limit 4KB)
		var bodyBuf []byte
		if c.Request.Body != nil {
			bodyBuf, _ = io.ReadAll(io.LimitReader(c.Request.Body, 4096))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBuf))
		}

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start).Milliseconds()
		userID := ""
		if v, ok := c.Get("authClaims"); ok {
			if claims, ok2 := v.(map[string]any); ok2 {
				if sub, ok3 := claims["sub"].(string); ok3 {
					userID = sub
				}
			}
		}
		traceID := ""
		if span := trace.SpanContextFromContext(c.Request.Context()); span.IsValid() {
			traceID = span.TraceID().String()
		}

		// mask headers
		headers := map[string][]string{}
		for k, v := range c.Request.Header {
			lk := strings.ToLower(k)
			if _, bad := sensitiveHeaders[lk]; bad {
				continue
			}
			headers[k] = v
		}

		// mask body fields (cheap: just omit if contains sensitive key)
		bodyStr := string(bodyBuf)
		truncated := len(bodyBuf) == 4096
		for _, key := range sensitiveKeys {
			if strings.Contains(strings.ToLower(bodyStr), key) {
				bodyStr = "[masked]"
				break
			}
		}

		lc.Enqueue(loki.LogRecord{
			Time:       time.Now(),
			Level:      levelFromStatus(status),
			Path:       c.FullPath(),
			Method:     c.Request.Method,
			Status:     status,
			DurationMs: duration,
			RequestID:  reqID,
			UserID:     userID,
			TraceID:    traceID,
			Message:    "http_request",
			URL:        c.Request.URL.String(),
			Query:      c.Request.URL.Query(),
			Headers:    headers,
			Body:       bodyStr,
			BodyTrunc:  truncated,
			RemoteIP:   c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
		})
	}
}

// levelFromStatus maps HTTP status code to log level
func levelFromStatus(code int) string {
	switch {
	case code >= http.StatusInternalServerError:
		return "error"
	case code >= http.StatusBadRequest:
		return "warn"
	default:
		return "info"
	}
}

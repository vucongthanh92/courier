package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/vucongthanh92/go-base-utils/logger"
	"github.com/vucongthanh92/go-base-utils/slack"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RecoverPanicMiddlewareConfig struct {
	SlackConfig slack.SlackConfig
}

// SafeGoMiddleware is a middleware that recovers from any panics
// and writes a 500 error response.
func RecoverPanicMiddleware(config RecoverPanicMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic
				logger.Error("Recovered from panic: ", zap.Any("panic", r), zap.String("stack", string(debug.Stack())))

				// stringutils.SendSlackMessage(" Https Server Panic: ", string(debug.Stack())
				slack.SendSlackMessage(config.SlackConfig, string(debug.Stack()))

				// Respond with a 500 Internal Server Error
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"errors":  "Internal Server Error",
					"success": false,
					"data":    nil,
				})
			}
		}()

		// Process request
		c.Next()
	}
}

package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// cors is a middleware that adds CORS headers to the response.
// It also handles preflight requests.
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS
// for more information about CORS.

// gin cors middleware

func Cors(allowOrigins ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		allows := strings.Join(allowOrigins, ",")
		if allows == "" {
			c.Next()
			return
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", allows)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

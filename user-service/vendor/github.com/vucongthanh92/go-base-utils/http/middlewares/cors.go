package middlewares

import (
	"net/http"
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

// CorsV2 is a middleware that adds CORS headers to the response.
// It also handles preflight requests.
func CorsV2(allowOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowOrigins))
	for _, origin := range allowOrigins {
		allowed[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
			c.Writer.Header().Add("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

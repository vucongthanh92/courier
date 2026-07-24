package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/helper/utils"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

func MessageRateLimitMiddleware(limiter interfaces.MessageRateLimiterI) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || conversationID == 0 {
			exposeRateLimitError(c, http.StatusBadRequest, "invalid_conversation_id", "conversation id must be a positive integer")
			return
		}

		claimsValue, ok := c.Get("authClaims")
		claims, claimsOK := claimsValue.(jwt.MapClaims)
		if !ok || !claimsOK {
			exposeRateLimitError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
			return
		}
		userID := utils.ParseUserID(claims["sub"])
		if userID == 0 {
			exposeRateLimitError(c, http.StatusUnauthorized, "unauthorized", "missing authenticated user")
			return
		}

		result, limitErr := limiter.Allow(c.Request.Context(), userID, conversationID)
		if limitErr != nil {
			logger.Warn("message rate limiter unavailable; allowing request",
				zap.Error(limitErr),
				zap.Uint64("user_id", userID),
				zap.Uint64("conversation_id", conversationID),
			)
			c.Next()
			return
		}

		if result.Limit > 0 {
			c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		}
		if result.Allowed {
			c.Next()
			return
		}

		c.Header("Retry-After", strconv.Itoa(result.RetryAfter))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt, 10))
		exposeRateLimitError(c, http.StatusTooManyRequests, "message_rate_limit_exceeded", "You are sending messages too quickly. Please try again later.")
	}
}

func exposeRateLimitError(c *gin.Context, status int, code, message string) {
	errHandler.InitErrorBuilder(c).
		SetStatus(status).
		SetError(models.ErrorDTO{Code: code, Message: message}).
		ExposeHttpError(c)
	c.Abort()
}

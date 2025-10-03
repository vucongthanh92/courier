package middlewares

import (
	"github.com/vucongthanh92/go-base-utils/logger"

	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.uber.org/zap"
)

func RateLimiting(logger logger.Logger, router *gin.Engine, rateFormat string) gin.HandlerFunc {
	router.ForwardedByClientIP = true
	rate, err := limiter.NewRateFromFormatted(rateFormat)
	if err != nil {
		logger.Fatal("RateLimiting", zap.Error(err))
		return nil
	}

	store := memory.NewStore()

	// Create a new middleware with the limiter instance.
	middleware := mgin.NewMiddleware(limiter.New(store, rate))

	return middleware
}

package errorhandler

import (
	"github.com/gin-gonic/gin"
	httpcommon "github.com/vucongthanh92/courier/user-service/helper/http_common"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

// ExposeHttpError sends the error response to the client using the Gin context.
// It constructs a standardized error response format and sets the appropriate HTTP status code.
func (b *ErrorBuilder) ExposeHttpError(c *gin.Context) {

	errors := []models.ErrorDTO{}

	utils.IterateSlice(b.Errors, func(i int, err models.ErrorDTO) {
		errors = append(errors, err)
	})

	response := httpcommon.SuccessResponse[any]{
		Success: false,
		Data:    nil,
		Errors:  errors,
	}

	c.JSON(b.Status, response)
}

// ExposeLogError logs the error details using the logger.
// It includes the main error, status code, and any additional errors if present.
func (b *ErrorBuilder) ExposeLogError() {
	if b == nil || b.LogError == nil {
		return
	}

	fields := []zap.Field{
		zap.Error(b.LogError),
		zap.Int("status", b.Status),
		zap.Bool("is_system_error", b.IsSystemError),
		zap.Bool("is_multiple_error", b.IsMultipleError),
		zap.Any("errors", b.Errors),
	}

	logger.Error("ErrorBuilder", fields...)
}

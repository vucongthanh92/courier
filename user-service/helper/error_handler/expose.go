package errorhandler

import (
	"github.com/gin-gonic/gin"
	httpcommon "github.com/vucongthanh92/courier/user-service/helper/http_common"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

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

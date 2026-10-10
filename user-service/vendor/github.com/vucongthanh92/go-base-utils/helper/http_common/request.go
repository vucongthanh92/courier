package httpcommon

import (
	"github.com/gin-gonic/gin"
	"github.com/vucongthanh92/go-base-utils/tracing"
)

func GetBodyParamsHTTP(c *gin.Context, dest interface{}) (err error) {
	_, span := tracing.StartSpanFromContext(c.Request.Context(), "GetBodyParamsHTTP")
	defer span.End()

	if err = c.ShouldBindJSON(&dest); err != nil {
		return
	}

	// time.ParseError

	return
}

func GetQueryParamsHTTP(c *gin.Context, dest interface{}) (err error) {
	_, span := tracing.StartSpanFromContext(c.Request.Context(), "GetQueryParamsHTTP")
	defer span.End()
	if err = c.ShouldBindQuery(dest); err != nil {
		return
	}
	return
}

package httpcommon

import (
	"github.com/gin-gonic/gin"
)

func GetBodyParamsHTTP(c *gin.Context, dest interface{}) (err error) {
	if err = c.ShouldBindJSON(&dest); err != nil {
		return
	}

	// time.ParseError

	return
}

func GetQueryParamsHTTP(c *gin.Context, dest interface{}) (err error) {
	if err = c.ShouldBindQuery(dest); err != nil {
		return
	}
	return
}

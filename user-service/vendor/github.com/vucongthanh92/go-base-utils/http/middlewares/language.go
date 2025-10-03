package middlewares

import (
	"github.com/vucongthanh92/go-base-utils/localization"

	"github.com/gin-gonic/gin"
)

func SetLanguage(resources []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Request.FormValue("lang")
		accept := c.Request.Header.Get("Accept-Language")
		localization.NewLocalizer(localization.ResourceConfig{
			Lang:      lang,
			Accept:    accept,
			Resources: resources,
		})
	}
}

package v1

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func MapRoutes(
	router *gin.Engine,
	userHandler *UserHandler,
	identityHandler *IdentityHandler,
) {
	v1 := router.Group("/api/v1")
	{
		// API for user
		v1.POST("/user/sign-up", userHandler.Signup)

		// API for identity
		v1.GET("/identity/create", identityHandler.CreateIdentity)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

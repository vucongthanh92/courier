package v1

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func MapRoutes(
	router *gin.Engine,
	authHandler *AuthHandler,
	identityHandler *IdentityHandler,
) {
	v1 := router.Group("/api/v1")
	{
		// API for auth
		v1.POST("/auth/sign-up", authHandler.Signup)

		// API for identity
		v1.POST("/auth/identity/create", identityHandler.CreateIdentity)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

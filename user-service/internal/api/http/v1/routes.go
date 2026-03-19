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
		v1.POST("/auth/verify-email", authHandler.VerifyEmail)
		v1.PUT("/auth/verify-email/resend", authHandler.ResendVerifyEmail)

		v1.POST("/auth/login", authHandler.Login)
		v1.POST("/auth/token/refresh", authHandler.RefreshToken)
		v1.POST("/auth/logout", authHandler.Logout)

		// API for identity
		v1.POST("/auth/identity/create", identityHandler.CreateIdentity)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

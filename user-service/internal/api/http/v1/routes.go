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
	authMiddleWare gin.HandlerFunc,
) {

	// Public routes
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/sign-up", authHandler.Signup)
		auth.POST("/verify-email", authHandler.VerifyEmail)
		auth.PUT("/verify-email/resend", authHandler.ResendVerifyEmail)
		auth.POST("/login", authHandler.Login)
		auth.PATCH("/refresh", authHandler.RefreshToken)

		auth.POST("/3rd/:provider", authHandler.OAuthLogin)

		// API for identity
		auth.POST("/identity/create", identityHandler.CreateIdentity)
	}

	// Protected routes
	v1 := router.Group("/api/v1")
	v1.Use(authMiddleWare)
	{
		v1.POST("/user/logout", authHandler.Logout)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

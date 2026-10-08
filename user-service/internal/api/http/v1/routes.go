package v1

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func MapRoutes(
	router *gin.Engine,
	authHandler *AuthHandler,
	credentialHandler *CredentialHandler,
	identityHandler *IdentityHandler,
	userHandler *UserHandler,
	ssoHandler *SsoHandler,
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

		// Routes for 3rd party OAuth login
		auth.POST("/identity/:provider", identityHandler.OAuthLogin)
		auth.GET("/identity/:provider/callback", identityHandler.OAuthCallback)
	}

	sso := router.Group("/api/v1/sso")
	{
		sso.GET("/authorize", ssoHandler.Authorize)
		sso.POST("/login", ssoHandler.Login)
		sso.POST("/token", ssoHandler.Token)
		sso.GET("/session", ssoHandler.Session)
		sso.POST("/logout", ssoHandler.Logout)
	}

	// Protected routes
	v1 := router.Group("/api/v1")
	v1.Use(authMiddleWare)
	{
		v1.POST("/user/logout", authHandler.Logout)
		v1.POST("/user/pwd/generate", credentialHandler.GeneratePassword)
		v1.GET("/user/search", userHandler.SearchUsers)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

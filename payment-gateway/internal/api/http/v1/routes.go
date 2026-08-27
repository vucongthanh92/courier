package v1

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func MapRoutes(
	router *gin.Engine,
	topUpHandler *TopUpHandler,
	sePayWebhookHandler *SePayWebhookHandler,
	authMiddleWare gin.HandlerFunc,
) {

	// path origin
	origin := router.Group("/api")

	// Provider webhooks authenticate with their provider signature, never a Courier JWT.
	v1Public := origin.Group("/v1")
	v1Public.POST("/webhooks/sepay", sePayWebhookHandler.Receive)

	// Protected REST routes
	v1HasAuth := origin.Group("/v1")
	v1HasAuth.Use(authMiddleWare)
	{
		v1HasAuth.POST("/wallet/top-ups", topUpHandler.Create)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

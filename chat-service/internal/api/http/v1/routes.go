package v1

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func MapRoutes(
	router *gin.Engine,
	conversationHandler *ConversationHandler,
	messageHandler *MessageHandler,
	authMiddleWare gin.HandlerFunc,
	messageRateLimitMiddleware gin.HandlerFunc,
) {

	// Protected routes
	v1 := router.Group("/api/v1")
	v1.Use(authMiddleWare)
	{
		v1.POST("/conversation/create", conversationHandler.CreateConversation)
		v1.POST("/conversations/:id/messages/create", messageRateLimitMiddleware, messageHandler.CreateMessage)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

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
	realtimeHandler *RealtimeHandler,
	authMiddleWare gin.HandlerFunc,
	messageRateLimitMiddleware gin.HandlerFunc,
) {

	v1 := router.Group("/api/v1")
	if realtimeHandler != nil {
		v1.GET("/ws", realtimeHandler.Handle)
	}

	// Protected REST routes
	protected := v1.Group("")
	protected.Use(authMiddleWare)
	{
		protected.POST("/conversation/create", conversationHandler.CreateConversation)
		protected.GET("/conversations", conversationHandler.ListConversations)
		protected.POST("/conversation/:id/messages/create", messageRateLimitMiddleware, messageHandler.CreateMessage)
		protected.GET("/conversation/:id/messages", messageHandler.ListMessages)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

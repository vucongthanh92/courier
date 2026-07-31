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
	wsHandler *WsHandler,
	authMiddleWare gin.HandlerFunc,
	messageRateLimitMiddleware gin.HandlerFunc,
) {

	// path origin
	origin := router.Group("/api")

	// non auth
	v1NonAuth := origin.Group("/v1")
	if wsHandler != nil {
		v1NonAuth.GET("/ws", wsHandler.VerifyAndConnect)
	}

	// Protected REST routes
	v1HasAuth := origin.Group("/v1")
	v1HasAuth.Use(authMiddleWare)
	{
		v1HasAuth.POST("/conversation/create", conversationHandler.CreateConversation)
		v1HasAuth.GET("/conversations", conversationHandler.ListConversations)
		v1HasAuth.POST("/conversation/:id/messages/create", messageRateLimitMiddleware, messageHandler.CreateMessage)
		v1HasAuth.GET("/conversation/:id/messages", messageHandler.ListMessages)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

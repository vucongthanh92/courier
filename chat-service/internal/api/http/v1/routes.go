package v1

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func MapRoutes(
	router *gin.Engine,
	conversationHandler *ConversationHandler,
) {

	// Public routes
	conversation := router.Group("/api/v1/conversations")
	{
		conversation.POST("/", conversationHandler.CreateConversation)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

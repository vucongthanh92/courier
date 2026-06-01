package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vucongthanh92/courier/chat-service/helper/utils"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
)

type ConversationHandler struct {
	conversationService interfaces.ConversationServiceI
}

func InitConversationHandler(
	conversationService interfaces.ConversationServiceI,
) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
	}
}

// API Create Conversation godoc
// @Tags Conversation
// @Summary create new conversation
// @Accept json
// @Produce json
// @Param params body models.CreateConversationRequest true "CreateConversationRequest"
// @Router /api/v1/conversations [post]
// @Success	200 {object} nil
func (h *ConversationHandler) CreateConversation(c *gin.Context) {

	// Call usecase
	ctx := utils.SetHeaderByKey(c, "headers")
	h.conversationService.CreateConversation(ctx)

	// Return response
	c.JSON(http.StatusOK, nil)
}

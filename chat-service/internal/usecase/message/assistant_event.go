package message

import (
	"context"
	"strconv"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	"github.com/vucongthanh92/courier/chat-service/helper/utils"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

func (s *messageUseCase) publishAssistantRequestedIfNeeded(ctx context.Context, conversation *entities.Conversation, message *entities.Message) {
	if s.assistantPublisher == nil || conversation == nil || message == nil {
		return
	}
	if conversation.Type != constants.ConversationTypeSystem || conversation.Name == nil || *conversation.Name != constants.SystemConversationAssistant {
		return
	}
	correlationID := "assistant-" + strconv.FormatUint(message.ID, 10) + "-" + utils.RandString(12)
	_ = s.assistantPublisher.PublishAssistantRequested(ctx, models.AssistantRequestedPayload{
		ConversationID:      message.ConversationID,
		TriggeringMessageID: message.ID,
		SenderID:            message.SenderID,
		Body:                message.Body,
		CorrelationID:       correlationID,
	})
}

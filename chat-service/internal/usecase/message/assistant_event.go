package message

import (
	"context"
	"strconv"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	"github.com/vucongthanh92/courier/chat-service/helper/utils"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

func (s *messageUseCase) publishAssistantRequestedIfNeeded(ctx context.Context, conversation *entities.Conversation, message *entities.Message) {
	if s.assistantPublisher == nil || conversation == nil || message == nil {
		logger.Warn("assistant request publisher is not configured")
		return
	}
	if conversation.Type != constants.ConversationTypeSystem || conversation.Name == nil || *conversation.Name != constants.SystemConversationAssistant {
		if conversation.Type == constants.ConversationTypeSystem {
			name := ""
			if conversation.Name != nil {
				name = *conversation.Name
			}
			logger.Info("skip assistant request publish for non-assistant system conversation", zap.Uint64("conversation_id", conversation.ID), zap.String("conversation_name", name))
		}
		return
	}
	correlationID := "assistant-" + strconv.FormatUint(message.ID, 10) + "-" + utils.RandString(12)
	payload := models.AssistantRequestedPayload{
		ConversationID:      message.ConversationID,
		TriggeringMessageID: message.ID,
		SenderID:            message.SenderID,
		Body:                message.Body,
		CorrelationID:       correlationID,
	}
	if err := s.assistantPublisher.PublishAssistantRequested(ctx, payload); err != nil {
		logger.Error("publish assistant request failed", zap.Error(err), zap.Uint64("conversation_id", message.ConversationID), zap.Uint64("message_id", message.ID), zap.String("correlation_id", correlationID))
		return
	}
	logger.Info("published assistant request", zap.Uint64("conversation_id", message.ConversationID), zap.Uint64("message_id", message.ID), zap.String("correlation_id", correlationID))
}

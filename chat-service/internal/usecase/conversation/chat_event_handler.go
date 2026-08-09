package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type ChatEventHandler struct {
	conversationReadRepo interfaces.ConversationQueryRepoI
	conversationCmdRepo  interfaces.ConversationCommandRepoI
	messageService       interfaces.MessageServiceI
}

func InitChatEventHandler(
	conversationReadRepo interfaces.ConversationQueryRepoI,
	conversationCmdRepo interfaces.ConversationCommandRepoI,
	messageService interfaces.MessageServiceI,
) *ChatEventHandler {
	return &ChatEventHandler{
		conversationReadRepo: conversationReadRepo,
		conversationCmdRepo:  conversationCmdRepo,
		messageService:       messageService,
	}
}

func (h *ChatEventHandler) HandleConversationCreated(ctx context.Context, body []byte) *errHandler.ErrorBuilder {
	var envelope models.IntegrationEventEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return errHandler.InitErrorBuilder(ctx).SetStatus(http.StatusBadRequest).SetLogError(err).SetError(models.ErrorDTO{
			Code:    "invalid_event",
			Message: "event envelope is invalid",
		})
	}
	if envelope.EventType != models.ConversationCreatedEventType || envelope.EventVersion != 1 {
		return errHandler.InitErrorBuilder(ctx).SetStatus(http.StatusBadRequest).SetError(models.ErrorDTO{
			Code:    "unsupported_event_type",
			Message: "unsupported chat event",
		})
	}

	var payload models.ConversationCreatedPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return errHandler.InitErrorBuilder(ctx).SetStatus(http.StatusBadRequest).SetLogError(err).SetError(models.ErrorDTO{
			Code:    "invalid_event_payload",
			Message: "conversation.created payload is invalid",
		})
	}

	inserted, processedErr := h.conversationCmdRepo.CreateProcessedEvent(ctx, envelope.EventID, models.ConversationCreatedV1)
	if processedErr != nil {
		return processedErr
	}
	if !inserted {
		return nil
	}

	for _, userID := range payload.MemberUserIDs {
		notificationConversation, queryErr := h.conversationReadRepo.GetSystemConversation(ctx, userID, constants.SystemConversationNotification)
		if queryErr != nil {
			return queryErr
		}
		if notificationConversation == nil {
			return errHandler.InitErrorBuilder(ctx).SetStatus(http.StatusNotFound).SetError(models.ErrorDTO{
				Code:    "notification_conversation_missing",
				Message: fmt.Sprintf("notification conversation does not exist for user %d", userID),
			})
		}

		_, _, createErr := h.messageService.CreateSystemMessage(ctx, &models.CreateSystemMessageRequest{
			ConversationID: notificationConversation.ID,
			Body:           "Bạn đã được thêm vào một cuộc trò chuyện mới",
			Metadata: map[string]any{
				"event_id":               envelope.EventID,
				"event_type":             models.ConversationCreatedV1,
				"source_conversation_id": payload.ConversationID,
				"notification_type":      "conversation_created",
			},
		})
		if createErr != nil {
			return createErr
		}
	}

	return nil
}

package conversation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type UserEventHandler struct {
	conversationService interfaces.ConversationServiceI
	messageService      interfaces.MessageServiceI
}

func InitUserEventHandler(
	conversationService interfaces.ConversationServiceI,
	messageService interfaces.MessageServiceI,
) *UserEventHandler {
	return &UserEventHandler{
		conversationService: conversationService,
		messageService:      messageService,
	}
}

func (h *UserEventHandler) HandleUserEmailVerified(ctx context.Context, body []byte) *errHandler.ErrorBuilder {
	var envelope models.IntegrationEventEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return errHandler.InitErrorBuilder(ctx).SetStatus(400).SetLogError(err).SetError(models.ErrorDTO{
			Code:    "invalid_event",
			Message: "event envelope is invalid",
		})
	}
	if envelope.EventType != "user.email_verified" || envelope.EventVersion != 1 {
		return errHandler.InitErrorBuilder(ctx).SetStatus(400).SetError(models.ErrorDTO{
			Code:    "unsupported_event_type",
			Message: "unsupported user event",
		})
	}

	var payload models.UserEmailVerifiedPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return errHandler.InitErrorBuilder(ctx).SetStatus(400).SetLogError(err).SetError(models.ErrorDTO{
			Code:    "invalid_event_payload",
			Message: "user.email_verified payload is invalid",
		})
	}

	ensureResp, ensureErr := h.conversationService.EnsureSystemConversations(ctx, &models.EnsureSystemConversationsRequest{
		UserID:    payload.UserID,
		EventID:   envelope.EventID,
		EventType: models.UserEmailVerifiedV1,
	})
	if ensureErr != nil {
		return ensureErr
	}
	if ensureResp == nil || !ensureResp.Processed {
		return nil
	}

	for _, item := range ensureResp.Conversations {
		if item.Name != nil && *item.Name == constants.SystemConversationNotification {
			_, _, createErr := h.messageService.CreateSystemMessage(ctx, &models.CreateSystemMessageRequest{
				ConversationID: item.ID,
				Body:           "Email verified successfully.",
				Metadata: map[string]any{
					"event_id":   envelope.EventID,
					"event_type": models.UserEmailVerifiedV1,
				},
			})
			return createErr
		}
	}

	return errHandler.InitErrorBuilder(ctx).SetStatus(500).SetError(models.ErrorDTO{
		Code:    "notification_conversation_missing",
		Message: fmt.Sprintf("notification conversation was not created for user %d", payload.UserID),
	})
}

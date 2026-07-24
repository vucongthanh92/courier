package message

import (
	"context"
	"net/http"
	"strings"

	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type UseCase struct {
	conversationQuery interfaces.ConversationQueryRepoI
	memberQuery       interfaces.MemberQueryRepoI
	messageQuery      interfaces.MessageQueryRepoI
	messageCommand    interfaces.MessageCmdRepoI
}

func InitMessageUsecase(
	conversationQuery interfaces.ConversationQueryRepoI,
	memberQuery interfaces.MemberQueryRepoI,
	messageQuery interfaces.MessageQueryRepoI,
	messageCommand interfaces.MessageCmdRepoI,
) interfaces.MessageServiceI {
	return &UseCase{
		conversationQuery: conversationQuery,
		memberQuery:       memberQuery,
		messageQuery:      messageQuery,
		messageCommand:    messageCommand,
	}
}

// func CreateMessage handles the creation of a new message in a conversation.
// It performs validation, checks for existing messages, and persists the new message if all checks pass.
func (s *UseCase) CreateMessage(ctx context.Context, req *models.SendMessageRequest) (*entities.Message, bool, *errHandler.ErrorBuilder) {

	// Validate the request body using the ValidateRequest method of SendMessageRequest
	if messageCode, messageErr := req.ValidateRequest(); messageCode != "" {
		return nil, false, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    messageCode,
				Message: messageErr,
			})
	}

	// Check if the conversation exists
	conversation, queryErr := s.conversationQuery.GetConversationByID(ctx, req.ConversationID)
	if queryErr != nil {
		return nil, false, queryErr
	}

	// If the conversation does not exist, return a 404 Not Found error
	if conversation == nil {
		return nil, false, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusNotFound).
			SetError(models.ErrorDTO{
				Code:    "conversation_not_found",
				Message: "conversation does not exist",
				Field:   "conversation_id",
			})
	}

	// Check if the sender is an active member of the conversation
	member, queryErr := s.memberQuery.GetConversationMember(ctx, req.ConversationID, req.SenderID)
	if queryErr != nil {
		return nil, false, queryErr
	}

	// If the member is not found or is not active, return a 403 Forbidden error
	if member == nil || member.Status != "active" {
		return nil, false, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusForbidden).
			SetError(models.ErrorDTO{
				Code:    "not_active_conversation_member",
				Message: "only active conversation members can send messages",
				Field:   "conversation_id",
			})
	}

	// Check for existing message with the same client_message_id if provided
	if req.ClientMessageID != nil {
		existing, existingErr := s.messageQuery.GetMessageByClientMessageID(ctx, req.ConversationID, *req.ClientMessageID)
		if existingErr != nil {
			return nil, false, existingErr
		}
		if existing != nil {
			if existing.SenderID != req.SenderID {
				return nil, false, errHandler.InitErrorBuilder(ctx).
					SetStatus(http.StatusForbidden).
					SetError(models.ErrorDTO{
						Code:    "client_message_id_conflict",
						Message: "client_message_id is already used by another sender",
						Field:   "client_message_id",
					})
			}
			return existing, false, nil
		}
	}

	// If reply_to_message_id is provided, check if the message exists and belongs to the same conversation
	if req.ReplyToMessageID != nil {
		reply, replyErr := s.messageQuery.GetMessageByID(ctx, *req.ReplyToMessageID)
		if replyErr != nil {
			return nil, false, replyErr
		}
		if reply == nil {
			return nil, false, errHandler.InitErrorBuilder(ctx).
				SetStatus(http.StatusNotFound).
				SetError(models.ErrorDTO{
					Code:    "reply_message_not_found",
					Message: "reply message does not exist",
					Field:   "reply_to_message_id",
				})
		}
		if reply.ConversationID != req.ConversationID {
			return nil, false, errHandler.InitErrorBuilder(ctx).
				SetStatus(http.StatusBadRequest).
				SetError(models.ErrorDTO{
					Code:    "reply_message_conversation_mismatch",
					Message: "reply message belongs to another conversation",
					Field:   "reply_to_message_id",
				})
		}
	}

	// Create a new message entity and persist it
	newMessageEntity := entities.Message{}
	if initErr := newMessageEntity.InitMessageEntity(
		req.ConversationID,
		req.SenderID,
		req.Type,
		req.Body,
		req.ClientMessageID,
		req.ReplyToMessageID,
		req.Metadata,
	); initErr != nil {
		return nil, false, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusInternalServerError).
			SetLogError(initErr).
			SetIsSystemError(true).
			SetError(models.ErrorDTO{Code: "system_error", Message: "There was an error on the server side"})
	}

	// Attempt to create the message in the database. If a unique constraint violation occurs,
	// check for an existing message with the same client_message_id and return it if found.
	if createErr := s.messageCommand.CreateMessage(ctx, &newMessageEntity); createErr != nil {
		if req.ClientMessageID != nil && isUniqueViolation(createErr.LogError) {
			existedMessage, existingErr := s.messageQuery.GetMessageByClientMessageID(ctx, req.ConversationID, *req.ClientMessageID)
			if existingErr != nil {
				return nil, false, existingErr
			}
			if existedMessage != nil && existedMessage.SenderID == req.SenderID {
				return existedMessage, false, nil
			}
		}
		return nil, false, createErr
	}

	return &newMessageEntity, true, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "sqlstate 23505") ||
		strings.Contains(message, "duplicate key")
}

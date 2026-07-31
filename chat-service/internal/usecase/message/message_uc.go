package message

import (
	"context"
	"net/http"
	"time"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type messageUseCase struct {
	conversationQuery interfaces.ConversationQueryRepoI
	memberQuery       interfaces.MemberQueryRepoI
	messageQuery      interfaces.MessageQueryRepoI
	messageCommand    interfaces.MessageCmdRepoI
	messageListCache  interfaces.MessageListCacheI
	wsPublisher       interfaces.WsPublisherI
}

func InitMessageUsecase(
	conversationQuery interfaces.ConversationQueryRepoI,
	memberQuery interfaces.MemberQueryRepoI,
	messageQuery interfaces.MessageQueryRepoI,
	messageCommand interfaces.MessageCmdRepoI,
	messageListCache interfaces.MessageListCacheI,
	wsPublisher interfaces.WsPublisherI,
) interfaces.MessageServiceI {
	return &messageUseCase{
		conversationQuery: conversationQuery,
		memberQuery:       memberQuery,
		messageQuery:      messageQuery,
		messageCommand:    messageCommand,
		messageListCache:  messageListCache,
		wsPublisher:       wsPublisher,
	}
}

// func CreateMessage handles the creation of a new message in a conversation.
// It performs validation, checks for existing messages, and persists the new message if all checks pass.
func (s *messageUseCase) CreateMessage(ctx context.Context, req *models.SendMessageRequest) (*models.MessageResponse, bool, *errHandler.ErrorBuilder) {

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

			var response models.MessageResponse
			response.MappeDTO(existing)
			return &response, false, nil
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
		if req.ClientMessageID != nil && createErr.IsUniqueViolation() {
			existedMessage, existingErr := s.messageQuery.GetMessageByClientMessageID(ctx, req.ConversationID, *req.ClientMessageID)
			if existingErr != nil {
				return nil, false, existingErr
			}
			if existedMessage != nil && existedMessage.SenderID == req.SenderID {
				var response models.MessageResponse
				response.MappeDTO(existedMessage)
				return &response, false, nil
			}
		}
		return nil, false, createErr
	}

	if s.messageListCache != nil {
		_ = s.messageListCache.InvalidateLatest(ctx, req.ConversationID)
	}

	// Map the newly created message entity to a MessageResponse and return it
	var response models.MessageResponse
	response.MappeDTO(&newMessageEntity)
	s.publishMessageCreated(ctx, req.ConversationID, response)
	return &response, true, nil
}

// ListMessages retrieves a list of messages in a conversation based on the provided request parameters.
// It performs validation, checks for conversation existence and membership, and returns the messages along with pagination information.
func (s *messageUseCase) ListMessages(ctx context.Context, req *models.ListMessagesRequest) (*models.ListMessagesResponse, *errHandler.ErrorBuilder) {

	if messageCode, messageErr := req.ValidateRequest(); messageCode != "" {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusBadRequest).
			SetError(models.ErrorDTO{
				Code:    messageCode,
				Message: messageErr,
			})
	}

	// Check if the conversation exists
	conversation, queryErr := s.conversationQuery.GetConversationByID(ctx, req.ConversationID)
	if queryErr != nil {
		return nil, queryErr
	}
	if conversation == nil {
		return nil, errHandler.InitErrorBuilder(ctx).
			SetStatus(http.StatusNotFound).
			SetError(models.ErrorDTO{
				Code:    "conversation_not_found",
				Message: "conversation does not exist",
				Field:   "conversation_id",
			})
	}

	// get list member of conversation
	members, memberErr := s.memberQuery.ListConversationMembers(ctx, req.ConversationID)
	if memberErr != nil {
		return nil, memberErr
	}
	memberResponses := make([]models.ConversationMemberResponse, 0, len(members))
	for i := range members {
		if members[i].UserID == req.RequesterID && members[i].Status != "active" {
			return nil, errHandler.InitErrorBuilder(ctx).
				SetStatus(http.StatusForbidden).
				SetError(models.ErrorDTO{
					Code:    "not_active_conversation_member",
					Message: "only active conversation members can read messages",
					Field:   "conversation_id",
				})
		}

		var mem models.ConversationMemberResponse
		mem.MappeDTO(members[i])
		memberResponses = append(memberResponses, mem)
	}

	//---
	if req.BeforeMessageID == nil && s.messageListCache != nil {
		cached, cacheErr := s.messageListCache.GetLatest(ctx, req.ConversationID, req.Limit)
		if cacheErr == nil && cached != nil {
			return &models.ListMessagesResponse{
				ConversationID: req.ConversationID,
				Messages:       cached.Messages,
				Members:        memberResponses,
				Pagination:     cached.Pagination,
			}, nil
		}
	}

	queryReq := *req
	queryReq.Limit = req.Limit + 1
	messages, listErr := s.messageQuery.ListMessages(ctx, req.ConversationID, queryReq)
	if listErr != nil {
		return nil, listErr
	}

	hasMore := len(messages) > req.Limit
	if hasMore {
		messages = messages[:req.Limit]
	}

	responseMessages := make([]models.MessageResponse, 0, len(messages))
	for i := range messages {
		el := models.MessageResponse{}
		el.MappeDTO(&messages[i])
		responseMessages = append(responseMessages, el)
	}

	// Determine the nextBeforeMessageID for pagination if there are more messages to fetch
	var nextBeforeMessageID *uint64
	if hasMore && len(messages) > 0 {
		lastID := messages[len(messages)-1].ID
		nextBeforeMessageID = &lastID
	}

	page := models.CachedMessageListPage{
		Messages: responseMessages,
		Pagination: models.MessagePaginationResponse{
			Limit:               req.Limit,
			NextBeforeMessageID: nextBeforeMessageID,
			HasMore:             hasMore,
		},
	}
	if req.BeforeMessageID == nil && s.messageListCache != nil {
		_ = s.messageListCache.SetLatest(ctx, req.ConversationID, req.Limit, page, constants.Time_Cache_15_minutes)
	}

	// Construct and return the ListMessagesResponse with the retrieved messages and pagination information
	return &models.ListMessagesResponse{
		ConversationID: req.ConversationID,
		Messages:       page.Messages,
		Members:        memberResponses,
		Pagination:     page.Pagination,
	}, nil
}

// private func publishMessageCreated
func (s *messageUseCase) publishMessageCreated(ctx context.Context, conversationID uint64, message models.MessageResponse) {
	if s.wsPublisher == nil {
		return
	}
	members, errBuilder := s.memberQuery.ListConversationMembers(ctx, conversationID)
	if errBuilder != nil {
		return
	}
	recipientIDs := make([]uint64, 0, len(members))
	for i := range members {
		if members[i].Status == "active" {
			recipientIDs = append(recipientIDs, members[i].UserID)
		}
	}
	if len(recipientIDs) == 0 {
		return
	}
	_ = s.wsPublisher.PublishMessageCreated(ctx, models.MessageCreatedEvent{
		Type:             models.MessageCreatedEventType,
		ConversationID:   conversationID,
		RecipientUserIDs: recipientIDs,
		Message:          message,
		EventAt:          time.Now().UTC(),
	})
}

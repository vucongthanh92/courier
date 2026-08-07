package models

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
)

// SendMessageRequest represents the request payload for sending a message in a conversation.
// It includes the message type, body, optional client message ID, optional reply-to message ID, and optional metadata.
type SendMessageRequest struct {
	Type             string         `json:"type" binding:"required"`
	Body             string         `json:"body" binding:"required"`
	ClientMessageID  *string        `json:"client_message_id,omitempty"`
	ReplyToMessageID *uint64        `json:"reply_to_message_id,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	SenderID         uint64         `json:"-"`
	ConversationID   uint64         `json:"-"`
}

type CreateSystemMessageRequest struct {
	ConversationID uint64
	Type           string
	Body           string
	Metadata       map[string]any
}

func (req *CreateSystemMessageRequest) ValidateRequest() (messageCode, messageErr string) {
	req.Body = strings.TrimSpace(req.Body)
	req.Type = strings.TrimSpace(req.Type)
	if req.Type == "" {
		req.Type = constants.MessageTypeSystem
	}
	switch {
	case req.ConversationID == 0:
		return "invalid_conversation_id", "conversation id must be a positive integer"
	case req.Type != constants.MessageTypeSystem && req.Type != constants.MessageTypeText:
		return "invalid_message_type", "system-created messages must be text or system"
	case req.Body == "":
		return "empty_message_body", "message body cannot be empty"
	case utf8.RuneCountInString(req.Body) > constants.MaxTextMessageRunes:
		return "message_body_too_long", "message body cannot exceed 4000 characters"
	case req.Metadata == nil:
		req.Metadata = map[string]any{}
	}
	return "", ""
}

func (req *SendMessageRequest) ValidateRequest() (messageCode, messageErr string) {
	req.Body = strings.TrimSpace(req.Body)
	switch {
	case req.SenderID == 0:
		return "unauthorized", "missing authenticated user"
	case req.ConversationID == 0:
		return "invalid_conversation_id", "conversation id must be a positive integer"
	case req.Type != "text":
		return "invalid_message_type", "only text messages are supported"
	case req.Body == "":
		return "empty_message_body", "message body cannot be empty"
	case utf8.RuneCountInString(req.Body) > constants.MaxTextMessageRunes:
		return "message_body_too_long", "message body cannot exceed 4000 characters"
	case req.ClientMessageID != nil:
		{
			value := strings.TrimSpace(*req.ClientMessageID)
			if value == "" || utf8.RuneCountInString(value) > constants.MaxClientMessageID {
				return "invalid_client_message_id", "client_message_id must contain 1 to 64 characters"
			}
			req.ClientMessageID = &value
		}
	case req.Metadata == nil:
		req.Metadata = map[string]any{}
	}
	return "", ""
}

// MessageResponse represents a message returned by chat APIs.
// It includes the message ID, conversation ID, sender ID, type, body, optional reply-to message ID, optional client message ID, metadata, and timestamps for creation, update, and edit.
type MessageResponse struct {
	ID               uint64         `json:"id,string"`
	ConversationID   uint64         `json:"conversation_id,string"`
	SenderID         uint64         `json:"sender_id,string"`
	Type             string         `json:"type"`
	Body             string         `json:"body"`
	ReplyToMessageID *uint64        `json:"reply_to_message_id,omitempty,string"`
	ClientMessageID  *string        `json:"client_message_id,omitempty"`
	Metadata         map[string]any `json:"metadata"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	EditedAt         *time.Time     `json:"edited_at,omitempty"`
}

func (m *MessageResponse) MappeDTO(message *entities.Message) {
	metadata := message.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	m.ID = message.ID
	m.ConversationID = message.ConversationID
	m.SenderID = message.SenderID
	m.Type = message.Type
	m.Body = message.Body
	m.ReplyToMessageID = message.ReplyToMessageID
	m.ClientMessageID = message.ClientMessageID
	m.Metadata = metadata
	m.CreatedAt = message.CreatedAt
	m.UpdatedAt = message.UpdatedAt
	m.EditedAt = message.EditedAt
}

// ListMessagesRequest represents the request parameters for listing messages in a conversation.
type ListMessagesRequest struct {
	Limit           int     `form:"limit"`
	BeforeMessageID *uint64 `form:"before_message_id"`
	RequesterID     uint64  `form:"-"`
	ConversationID  uint64  `form:"-"`
}

func (req *ListMessagesRequest) ValidateRequest() (messageCode, messageErr string) {
	switch {
	case req.RequesterID == 0:
		return "unauthorized", "missing authenticated user"
	case req.ConversationID == 0:
		return "invalid_conversation_id", "conversation id must be a positive integer"
	case req.Limit < 0:
		return "invalid_limit", "limit must be a positive integer"
	}
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	return "", ""
}

// ListMessagesResponse represents the response for listing messages in a conversation.
type ListMessagesResponse struct {
	ConversationID uint64                       `json:"conversation_id,string"`
	Messages       []MessageResponse            `json:"messages"`
	Members        []ConversationMemberResponse `json:"members"`
	Pagination     MessagePaginationResponse    `json:"pagination"`
}

// MessagePaginationResponse represents the pagination information for listing messages in a conversation.
type MessagePaginationResponse struct {
	Limit               int     `json:"limit"`
	NextBeforeMessageID *uint64 `json:"next_before_message_id,omitempty,string"`
	HasMore             bool    `json:"has_more"`
}

type CachedMessageListPage struct {
	Messages   []MessageResponse         `json:"messages"`
	Pagination MessagePaginationResponse `json:"pagination"`
}

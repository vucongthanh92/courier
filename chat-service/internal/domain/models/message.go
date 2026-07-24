package models

import (
	"strings"
	"unicode/utf8"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
)

// SendMessageRequest represents the request payload for sending a message in a conversation.
type SendMessageRequest struct {
	Type             string         `json:"type" binding:"required"`
	Body             string         `json:"body" binding:"required"`
	ClientMessageID  *string        `json:"client_message_id,omitempty"`
	ReplyToMessageID *uint64        `json:"reply_to_message_id,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	SenderID         uint64         `json:"-"`
	ConversationID   uint64         `json:"-"`
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

// ListMessagesRequest represents the request parameters for listing messages in a conversation.
type ListMessagesRequest struct {
	Limit           int     `form:"limit"`
	BeforeMessageID *uint64 `form:"before_message_id"`
}

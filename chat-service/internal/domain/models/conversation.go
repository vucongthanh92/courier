package models

import (
	"database/sql"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
)

// CreateConversationRequest represents the request payload for creating a new conversation.
type CreateConversationRequest struct {
	Type          string       `json:"type,omitempty" binding:"omitempty,oneof=direct group"`
	Name          *string      `json:"name,omitempty"`
	MemberUserIDs []UserIDJSON `json:"member_user_ids" binding:"required,min=1"`
	CreatorID     uint64
}

type UserIDJSON uint64

var _ json.Unmarshaler = (*UserIDJSON)(nil)
var _ encoding.TextUnmarshaler = (*UserIDJSON)(nil)

func (id *UserIDJSON) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	raw = strings.Trim(raw, `"`)
	return id.parse(raw)
}

func (id *UserIDJSON) UnmarshalText(text []byte) error {
	return id.parse(string(text))
}

func (id *UserIDJSON) parse(raw string) error {
	value, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user id %q: %w", raw, err)
	}
	*id = UserIDJSON(value)
	return nil
}

func (c *CreateConversationRequest) MemberIDs() []uint64 {
	memberIDs := make([]uint64, 0, len(c.MemberUserIDs))
	for _, userID := range c.MemberUserIDs {
		memberIDs = append(memberIDs, uint64(userID))
	}
	return memberIDs
}

func (c *CreateConversationRequest) ValidateConversationType(sortedMemberIDs []uint64) error {
	switch {
	case len(sortedMemberIDs) < 2:
		return errors.New("conversation must have at least 2 members including creator")
	case len(sortedMemberIDs) == 2:
		c.Type = constants.ConversationTypeDirect
	default:
		c.Type = constants.ConversationTypeGroup
	}
	return nil
}

// CreateConversationResponse represents the response payload after successfully creating a new conversation.
type CreateConversationResponse struct {
	ID            uint64                        `json:"id,string"`
	Type          string                        `json:"type"`
	DirectKey     *string                       `json:"direct_key,omitempty"`
	Name          *string                       `json:"name,omitempty"`
	CreatedBy     uint64                        `json:"created_by,string"`
	LastMessageID *uint64                       `json:"last_message_id,omitempty,string"`
	LastMessageAt *time.Time                    `json:"last_message_at,omitempty"`
	Metadata      map[string]any                `json:"metadata"`
	CreatedAt     time.Time                     `json:"created_at"`
	UpdatedAt     time.Time                     `json:"updated_at"`
	Members       []entities.ConversationMember `json:"members,omitempty"`
}

func (c *CreateConversationResponse) FromEntity(conversation *entities.Conversation, members []entities.ConversationMember) {
	c.ID = conversation.ID
	c.Type = conversation.Type
	c.DirectKey = conversation.DirectKey
	c.Name = conversation.Name
	c.CreatedBy = conversation.CreatedBy
	c.LastMessageID = conversation.LastMessageID
	c.LastMessageAt = conversation.LastMessageAt
	c.Metadata = conversation.Metadata
	c.CreatedAt = conversation.CreatedAt
	c.UpdatedAt = conversation.UpdatedAt
	c.Members = members
}

type ListConversationsRequest struct {
	Limit                int        `form:"limit"`
	BeforeLastMessageAt  *time.Time `form:"before_last_message_at" time_format:"2006-01-02T15:04:05Z07:00"`
	BeforeConversationID *uint64    `form:"before_conversation_id"`
	RequesterID          uint64     `form:"-"`
}

func (req *ListConversationsRequest) ValidateRequest() (messageCode, messageErr string) {
	switch {
	case req.RequesterID == 0:
		return "unauthorized", "missing authenticated user"
	case req.Limit < 0:
		return "invalid_limit", "limit must be a positive integer"
	}
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if (req.BeforeLastMessageAt == nil) != (req.BeforeConversationID == nil) {
		return "invalid_cursor", "before_last_message_at and before_conversation_id must be provided together"
	}
	return "", ""
}

// struct ConversationListItemResponse
type ConversationListResponse struct {
	ID            uint64           `json:"id,string"`
	Type          string           `json:"type"`
	DirectKey     string           `json:"direct_key,omitempty"`
	Name          string           `json:"name,omitempty"`
	CreatedBy     uint64           `json:"created_by,string"`
	LastMessageID *uint64          `json:"last_message_id,omitempty,string"`
	LastMessageAt *time.Time       `json:"last_message_at,omitempty"`
	Metadata      map[string]any   `json:"metadata"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	LastMessage   *MessageResponse `json:"last_message,omitempty"`
}

func (m *ConversationListResponse) MapConversationFromDB(row ConversationWithLastMessage) {
	metadata := row.DecodeJSONMap(row.Metadata)
	if row.DirectKey.Valid {
		m.DirectKey = row.DirectKey.String
	}
	if row.Name.Valid {
		m.Name = row.Name.String
	}

	m.ID = row.ID
	m.Type = row.Type
	m.CreatedBy = row.CreatedBy
	m.LastMessageID = row.LastMessageID
	m.LastMessageAt = row.LastMessageAt
	m.Metadata = metadata
	m.CreatedAt = row.CreatedAt
	m.UpdatedAt = row.UpdatedAt

	if row.MessageID != nil && row.MessageConversationID != nil &&
		row.MessageSenderID != nil && row.MessageType != nil && row.MessageBody != nil {
		messageMetadata := row.DecodeJSONMap(row.MessageMetadata)
		m.LastMessage = &MessageResponse{
			ID:               *row.MessageID,
			ConversationID:   *row.MessageConversationID,
			SenderID:         *row.MessageSenderID,
			Type:             *row.MessageType,
			Body:             *row.MessageBody,
			ReplyToMessageID: row.MessageReplyToID,
			ClientMessageID:  row.MessageClientID,
			Metadata:         messageMetadata,
			EditedAt:         row.MessageEditedAt,
		}
		if row.MessageCreatedAt != nil {
			m.LastMessage.CreatedAt = *row.MessageCreatedAt
		}
		if row.MessageUpdatedAt != nil {
			m.LastMessage.UpdatedAt = *row.MessageUpdatedAt
		}
	}
}

// struct ListConversationsPaginationResponse
type ListConversationsPaginationResponse struct {
	Limit                    int        `json:"limit"`
	HasMore                  bool       `json:"has_more"`
	NextBeforeLastMessageAt  *time.Time `json:"next_before_last_message_at,omitempty"`
	NextBeforeConversationID *uint64    `json:"next_before_conversation_id,omitempty,string"`
}

type ListConversationsResponse struct {
	Conversations []ConversationListResponse          `json:"conversations"`
	Pagination    ListConversationsPaginationResponse `json:"pagination"`
}

type EnsureSystemConversationsRequest struct {
	UserID    uint64
	EventID   string
	EventType string
}

func (req *EnsureSystemConversationsRequest) ValidateRequest() (messageCode, messageErr string) {
	req.EventID = strings.TrimSpace(req.EventID)
	req.EventType = strings.TrimSpace(req.EventType)
	switch {
	case req.UserID == 0:
		return "invalid_user_id", "user_id must be a positive integer"
	case req.EventID == "":
		return "invalid_event_id", "event_id is required"
	case req.EventType == "":
		return "invalid_event_type", "event_type is required"
	}
	return "", ""
}

type EnsureSystemConversationsResponse struct {
	UserID        uint64                       `json:"user_id,string"`
	Conversations []CreateConversationResponse `json:"conversations"`
	Processed     bool                         `json:"processed"`
}

func SystemConversationNames() []string {
	return []string{
		constants.SystemConversationNotification,
		constants.SystemConversationAssistant,
	}
}

// struct ConversationWithLastMessage
type ConversationWithLastMessage struct {
	ID                    uint64          `gorm:"column:id"`
	Type                  string          `gorm:"column:type"`
	DirectKey             sql.NullString  `gorm:"column:direct_key"`
	Name                  sql.NullString  `gorm:"column:name"`
	CreatedBy             uint64          `gorm:"column:created_by"`
	LastMessageID         *uint64         `gorm:"column:last_message_id"`
	LastMessageAt         *time.Time      `gorm:"column:last_message_at"`
	Metadata              json.RawMessage `gorm:"column:metadata"`
	CreatedAt             time.Time       `gorm:"column:created_at"`
	UpdatedAt             time.Time       `gorm:"column:updated_at"`
	MessageID             *uint64         `gorm:"column:message_id"`
	MessageConversationID *uint64         `gorm:"column:message_conversation_id"`
	MessageSenderID       *uint64         `gorm:"column:message_sender_id"`
	MessageType           *string         `gorm:"column:message_type"`
	MessageBody           *string         `gorm:"column:message_body"`
	MessageReplyToID      *uint64         `gorm:"column:message_reply_to_message_id"`
	MessageClientID       *string         `gorm:"column:message_client_message_id"`
	MessageMetadata       json.RawMessage `gorm:"column:message_metadata"`
	MessageCreatedAt      *time.Time      `gorm:"column:message_created_at"`
	MessageUpdatedAt      *time.Time      `gorm:"column:message_updated_at"`
	MessageEditedAt       *time.Time      `gorm:"column:message_edited_at"`
}

func (m *ConversationWithLastMessage) DecodeJSONMap(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var res map[string]any
	if err := json.Unmarshal(raw, &res); err != nil || res == nil {
		return map[string]any{}
	}
	return res
}

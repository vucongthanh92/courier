package models

import (
	"time"

	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
)

// CreateConversationRequest represents the request payload for creating a new conversation.
type CreateConversationRequest struct {
	Type          string   `json:"type" binding:"required,oneof=direct group"`
	Name          *string  `json:"name,omitempty"`
	MemberUserIDs []uint64 `json:"member_user_ids" binding:"required,min=2"`
	CreatorID     uint64
}

// CreateConversationResponse represents the response payload after successfully creating a new conversation.
type CreateConversationResponse struct {
	ID            uint64                        `json:"id"`
	Type          string                        `json:"type"`
	DirectKey     *string                       `json:"direct_key,omitempty"`
	Name          *string                       `json:"name,omitempty"`
	CreatedBy     uint64                        `json:"created_by"`
	LastMessageID *uint64                       `json:"last_message_id,omitempty"`
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

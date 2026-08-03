package models

import (
	"time"

	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
)

type ConversationMemberResponse struct {
	ID                uint64                     `json:"id,string"`
	ConversationID    uint64                     `json:"conversation_id,string"`
	UserID            uint64                     `json:"user_id,string"`
	Role              string                     `json:"role"`
	Status            string                     `json:"status"`
	JoinedAt          time.Time                  `json:"joined_at"`
	LeftAt            *time.Time                 `json:"left_at,omitempty"`
	LastReadMessageID *uint64                    `json:"last_read_message_id,omitempty,string"`
	LastReadAt        *time.Time                 `json:"last_read_at,omitempty"`
	MutedUntil        *time.Time                 `json:"muted_until,omitempty"`
	Profile           UserProfileSummaryResponse `json:"profile"`
}

func (m *ConversationMemberResponse) MappeDTO(entity entities.ConversationMember) {
	m.ID = entity.ID
	m.ConversationID = entity.ConversationID
	m.UserID = entity.UserID
	m.Role = entity.Role
	m.Status = entity.Status
	m.JoinedAt = entity.JoinedAt
	m.LeftAt = entity.LeftAt
	m.LastReadMessageID = entity.LastReadMessageID
	m.LastReadAt = entity.LastReadAt
	m.MutedUntil = entity.MutedUntil
}

type UserProfileSummaryResponse struct {
	UserID      uint64 `json:"user_id,string"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Status      string `json:"status"`
}

// ListConversationMembersRequest represents the request to list members of a conversation.
type ListConversationMembersRequest struct {
	ConversationID uint64 `uri:"id" binding:"required"`
	RequesterID    uint64 `uri:"-"`
}

func (req *ListConversationMembersRequest) ValidateRequest() (messageCode, messageErr string) {
	switch {
	case req.RequesterID == 0:
		return "unauthorized", "missing authenticated user"
	case req.ConversationID == 0:
		return "invalid_conversation_id", "conversation id must be a positive integer"
	}
	return "", ""
}

// ListConversationMembersResponse represents the response containing the list of members for a conversation.
type ListConversationMembersResponse struct {
	ConversationID uint64                       `json:"conversation_id,string"`
	Members        []ConversationMemberResponse `json:"members"`
}

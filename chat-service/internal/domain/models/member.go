package models

import (
	"time"

	"github.com/vucongthanh92/courier/chat-service/internal/domain/entities"
)

type ConversationMemberResponse struct {
	ID                uint64     `json:"id,string"`
	ConversationID    uint64     `json:"conversation_id,string"`
	UserID            uint64     `json:"user_id,string"`
	Role              string     `json:"role"`
	Status            string     `json:"status"`
	JoinedAt          time.Time  `json:"joined_at"`
	LeftAt            *time.Time `json:"left_at,omitempty"`
	LastReadMessageID *uint64    `json:"last_read_message_id,omitempty,string"`
	LastReadAt        *time.Time `json:"last_read_at,omitempty"`
	MutedUntil        *time.Time `json:"muted_until,omitempty"`
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

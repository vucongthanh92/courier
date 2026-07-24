package models

import (
	"time"
)

type ConversationMemberResponse struct {
	ID                uint64     `json:"id"`
	ConversationID    uint64     `json:"conversation_id"`
	UserID            uint64     `json:"user_id"`
	Role              string     `json:"role"`
	Status            string     `json:"status"`
	JoinedAt          time.Time  `json:"joined_at"`
	LeftAt            *time.Time `json:"left_at,omitempty"`
	LastReadMessageID *uint64    `json:"last_read_message_id,omitempty"`
	LastReadAt        *time.Time `json:"last_read_at,omitempty"`
	MutedUntil        *time.Time `json:"muted_until,omitempty"`
}

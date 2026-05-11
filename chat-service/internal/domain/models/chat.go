package models

import "time"

type CreateDirectConversationRequest struct {
	PeerUserID uint64 `json:"peer_user_id" binding:"required"`
}

type CreateGroupConversationRequest struct {
	Name          string   `json:"name" binding:"required"`
	MemberUserIDs []uint64 `json:"member_user_ids" binding:"required,min=1"`
}

type SendMessageRequest struct {
	Body             string  `json:"body" binding:"required"`
	ClientMessageID  *string `json:"client_message_id,omitempty"`
	ReplyToMessageID *uint64 `json:"reply_to_message_id,omitempty"`
}

type MarkConversationReadRequest struct {
	LastReadMessageID uint64 `json:"last_read_message_id" binding:"required"`
}

type ListMessagesRequest struct {
	Limit           int     `form:"limit"`
	BeforeMessageID *uint64 `form:"before_message_id"`
}

type ConversationResponse struct {
	ID            uint64              `json:"id"`
	Type          string              `json:"type"`
	DirectKey     *string             `json:"direct_key,omitempty"`
	Name          *string             `json:"name,omitempty"`
	CreatedBy     uint64              `json:"created_by"`
	LastMessageID *uint64             `json:"last_message_id,omitempty"`
	LastMessageAt *time.Time          `json:"last_message_at,omitempty"`
	Metadata      map[string]any      `json:"metadata"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	Members       []ConversationMemberResponse `json:"members,omitempty"`
}

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

type MessageResponse struct {
	ID               uint64         `json:"id"`
	ConversationID   uint64         `json:"conversation_id"`
	SenderID         uint64         `json:"sender_id"`
	Type             string         `json:"type"`
	Body             string         `json:"body"`
	ReplyToMessageID *uint64        `json:"reply_to_message_id,omitempty"`
	ClientMessageID  *string        `json:"client_message_id,omitempty"`
	Metadata         map[string]any `json:"metadata"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	EditedAt         *time.Time     `json:"edited_at,omitempty"`
	DeletedAt        *time.Time     `json:"deleted_at,omitempty"`
}

type InboxConversationResponse struct {
	Conversation ConversationResponse  `json:"conversation"`
	LastMessage  *MessageResponse      `json:"last_message,omitempty"`
	Member       ConversationMemberResponse `json:"member"`
	UnreadCount  int64                 `json:"unread_count"`
}

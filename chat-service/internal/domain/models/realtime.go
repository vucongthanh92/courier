package models

import "time"

const (
	RealtimeEventMessageCreated = "message.created"
)

type MessageCreatedEvent struct {
	Type             string          `json:"type"`
	ConversationID   uint64          `json:"conversation_id,string"`
	RecipientUserIDs []uint64        `json:"recipient_user_ids"`
	Message          MessageResponse `json:"message"`
	EventAt          time.Time       `json:"event_at"`
}

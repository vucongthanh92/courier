package models

import (
	"encoding/json"
	"time"
)

const (
	MessageCreatedEventType = "message.created"
	UserEmailVerifiedV1     = "user.email_verified.v1"
)

type MessageCreatedEvent struct {
	Type             string          `json:"type"`
	ConversationID   uint64          `json:"conversation_id,string"`
	RecipientUserIDs []uint64        `json:"recipient_user_ids"`
	Message          MessageResponse `json:"message"`
	EventAt          time.Time       `json:"event_at"`
}

type IntegrationEventEnvelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	EventVersion  int             `json:"event_version"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Source        string          `json:"source"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
}

type UserEmailVerifiedPayload struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

package models

import (
	"encoding/json"
	"time"
)

const (
	MessageCreatedEventType      = "message.created"
	ConversationCreatedEventType = "conversation.created"
	UserEmailVerifiedV1          = "user.email_verified.v1"
	ConversationCreatedV1        = "conversation.created.v1"
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

type ConversationCreatedPayload struct {
	ConversationID   uint64   `json:"conversation_id,string"`
	ConversationType string   `json:"conversation_type"`
	CreatedBy        uint64   `json:"created_by,string"`
	MemberUserIDs    []uint64 `json:"member_user_ids"`
}

type AssistantRequestedPayload struct {
	ConversationID      uint64 `json:"conversation_id"`
	TriggeringMessageID uint64 `json:"triggering_message_id"`
	SenderID            uint64 `json:"sender_id"`
	Body                string `json:"body"`
	CorrelationID       string `json:"correlation_id"`
}

type AssistantRespondedPayload struct {
	ConversationID      uint64                 `json:"conversation_id"`
	TriggeringMessageID uint64                 `json:"triggering_message_id"`
	Body                string                 `json:"body,omitempty"`
	MessageParts        []AssistantMessagePart `json:"message_parts,omitempty"`
	CorrelationID       string                 `json:"correlation_id"`
	Metadata            map[string]any         `json:"metadata,omitempty"`
}

type AssistantMessagePart struct {
	Body     string         `json:"body"`
	Index    int            `json:"index"`
	Total    int            `json:"total"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

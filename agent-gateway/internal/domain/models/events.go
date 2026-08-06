package models

import "time"

type EventEnvelope[T any] struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	EventVersion  int       `json:"event_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	Source        string    `json:"source"`
	AggregateType string    `json:"aggregate_type"`
	AggregateID   string    `json:"aggregate_id"`
	Payload       T         `json:"payload"`
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

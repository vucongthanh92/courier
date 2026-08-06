package models

import "time"

type MemoryPoint struct {
	ID             string         `json:"id"`
	ConversationID uint64         `json:"conversation_id"`
	ChatMessageID  uint64         `json:"chat_message_id,omitempty"`
	Role           string         `json:"role"`
	Body           string         `json:"body"`
	Source         string         `json:"source"`
	CorrelationID  string         `json:"correlation_id,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

type ContextPackage struct {
	SystemInstructions  string        `json:"system_instructions"`
	ConversationSummary string        `json:"conversation_summary,omitempty"`
	RecentMessages      []MemoryPoint `json:"recent_messages,omitempty"`
	RelevantMemories    []MemoryPoint `json:"relevant_memories,omitempty"`
	CurrentMessage      string        `json:"current_message"`
}

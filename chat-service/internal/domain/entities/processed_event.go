package entities

import "time"

type ProcessedEvent struct {
	EventID     string    `gorm:"column:event_id;primaryKey;type:text" json:"event_id"`
	EventType   string    `gorm:"column:event_type;type:text;not null" json:"event_type"`
	ProcessedAt time.Time `gorm:"column:processed_at;type:timestamptz;autoCreateTime" json:"processed_at"`
}

func (ProcessedEvent) TableName() string {
	return `"chat-service".processed_events`
}

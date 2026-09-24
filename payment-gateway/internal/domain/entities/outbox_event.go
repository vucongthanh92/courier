package entities

import (
	"strconv"
	"time"
)

type OutboxEvent struct {
	ID            uint64     `gorm:"column:id;primaryKey"`
	AggregateType string     `gorm:"column:aggregate_type"`
	AggregateID   string     `gorm:"column:aggregate_id"`
	EventType     string     `gorm:"column:event_type"`
	Payload       []byte     `gorm:"column:payload;type:jsonb"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
	Attempts      int        `gorm:"column:attempts"`
	LastError     *string    `gorm:"column:last_error"`
}

func (OutboxEvent) TableName() string {
	return `"payment-gateway".outbox_events`
}

func (e *OutboxEvent) Initialize(
	ID uint64,
	aggregateType string,
	aggregateID uint64,
	eventType string,
	payload []byte,
) {
	e.ID = ID
	e.AggregateType = aggregateType
	e.AggregateID = strconv.FormatUint(aggregateID, 10)
	e.EventType = eventType
	e.Payload = payload
}

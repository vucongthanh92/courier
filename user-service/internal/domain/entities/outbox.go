package entities

import "time"

// Outbox message for reliable publish (transactional outbox)
// Create records in the same tx as domain changes, then a worker will dispatch.
type Outbox struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement:false;check:id>0"`
	AggregateType string     `gorm:"type:text;not null"`
	AggregateID   string     `gorm:"type:text;not null"`
	EventType     string     `gorm:"type:text;not null"`
	Payload       []byte     `gorm:"type:bytea;not null"` // JSON bytes
	CreatedAt     time.Time  `gorm:"type:timestamptz;autoCreateTime"`
	PublishedAt   *time.Time `gorm:"type:timestamptz"`
}

func (Outbox) TableName() string { return "outbox" }

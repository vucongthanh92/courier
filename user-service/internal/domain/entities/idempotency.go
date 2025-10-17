package entities

import "time"

type Idempotency struct {
	Key        string    `gorm:"primaryKey;type:text"`
	RequestSig string    `gorm:"type:text;not null"`  // hash of normalized request body
	Response   []byte    `gorm:"type:bytea;not null"` // cached response JSON
	Status     string    `gorm:"type:text;not null"`  // pending|succeeded|failed
	CreatedAt  time.Time `gorm:"type:timestamptz;autoCreateTime"`
	ExpiresAt  time.Time `gorm:"type:timestamptz;index"`
}

func (Idempotency) TableName() string { return "idempotency" }

package entities

import "time"

type IdempotencyKey struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	Scope          string    `gorm:"column:scope"`
	UserID         *uint64   `gorm:"column:user_id"`
	IdempotencyKey string    `gorm:"column:idempotency_key"`
	RequestHash    string    `gorm:"column:request_hash"`
	ResponseStatus *int16    `gorm:"column:response_status"`
	ResponseBody   []byte    `gorm:"column:response_body;type:jsonb"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	ExpiresAt      time.Time `gorm:"column:expires_at"`
}

func (IdempotencyKey) TableName() string {
	return `"payment-gateway".idempotency_keys`
}

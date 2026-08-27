package entities

import "time"

type ProviderEvent struct {
	ID              uint64     `gorm:"column:id;primaryKey"`
	Provider        string     `gorm:"column:provider"`
	ProviderEventID string     `gorm:"column:provider_event_id"`
	Payload         []byte     `gorm:"column:payload;type:jsonb"`
	SignatureValid  bool       `gorm:"column:signature_valid"`
	Status          string     `gorm:"column:status"`
	ReceivedAt      time.Time  `gorm:"column:received_at"`
	ProcessedAt     *time.Time `gorm:"column:processed_at"`
	ErrorCode       *string    `gorm:"column:error_code"`
}

func (ProviderEvent) TableName() string { return `"payment-gateway".provider_events` }

type ProviderTransaction struct {
	ID                    uint64     `gorm:"column:id;primaryKey"`
	Provider              string     `gorm:"column:provider"`
	ProviderTransactionID string     `gorm:"column:provider_transaction_id"`
	TopUpIntentID         uint64     `gorm:"column:topup_intent_id"`
	AmountMinor           int64      `gorm:"column:amount_minor"`
	Currency              string     `gorm:"column:currency"`
	PaidAt                *time.Time `gorm:"column:paid_at"`
	ReceivingAccountKey   *string    `gorm:"column:receiving_account_key"`
	SourceMetadata        []byte     `gorm:"column:source_metadata;type:jsonb"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
}

func (ProviderTransaction) TableName() string { return `"payment-gateway".provider_transactions` }

package entities

import (
	"strconv"
	"time"
)

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

func (ProviderTransaction) TableName() string {
	return `"payment-gateway".provider_transactions`
}

func (e *ProviderTransaction) Initialize(
	ID uint64,
	providerName string,
	providerTransactionID int64,
	topupIntentID uint64,
	amountMinor int64,
	currency string,
	paidAt *time.Time,
	receivingAccountKey *string,
	sourceMetadata []byte,
) {
	e.ID = ID
	e.Provider = providerName
	e.ProviderTransactionID = strconv.FormatInt(providerTransactionID, 10)
	e.TopUpIntentID = topupIntentID
	e.AmountMinor = amountMinor
	e.Currency = currency
	e.PaidAt = paidAt
	e.ReceivingAccountKey = receivingAccountKey
	e.SourceMetadata = sourceMetadata
}

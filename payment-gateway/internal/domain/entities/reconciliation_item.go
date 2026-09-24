package entities

import "time"

type ReconciliationItem struct {
	ID                    uint64    `gorm:"column:id;primaryKey"`
	ReconciliationRunID   uint64    `gorm:"column:reconciliation_run_id"`
	ProviderTransactionID *string   `gorm:"column:provider_transaction_id"`
	TopUpIntentID         *uint64   `gorm:"column:topup_intent_id"`
	Status                string    `gorm:"column:status"`
	Detail                []byte    `gorm:"column:detail;type:jsonb"`
	CreatedAt             time.Time `gorm:"column:created_at"`
}

func (ReconciliationItem) TableName() string {
	return `"payment-gateway".reconciliation_items`
}

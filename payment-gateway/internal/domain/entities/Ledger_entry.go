package entities

import "time"

type LedgerEntry struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	JournalID   uint64    `gorm:"column:journal_id"`
	AccountID   uint64    `gorm:"column:account_id"`
	Side        string    `gorm:"column:side"`
	AmountMinor int64     `gorm:"column:amount_minor"`
	Currency    string    `gorm:"column:currency"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (LedgerEntry) TableName() string {
	return `"payment-gateway".ledger_entries`
}

func (e *LedgerEntry) Initialize(
	debitID uint64,
	journalID uint64,
	accountID uint64,
	side string,
	amountMinor int64,
	currency string,
) {
	e.ID = debitID
	e.JournalID = journalID
	e.AccountID = accountID
	e.Side = side
	e.AmountMinor = amountMinor
	e.Currency = currency
}

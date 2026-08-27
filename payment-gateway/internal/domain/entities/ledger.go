package entities

import "time"

type LedgerAccount struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	AccountCode string    `gorm:"column:account_code"`
	AccountType string    `gorm:"column:account_type"`
	Currency    string    `gorm:"column:currency"`
	WalletID    *uint64   `gorm:"column:wallet_id"`
	NormalSide  string    `gorm:"column:normal_side"`
	IsActive    bool      `gorm:"column:is_active"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (LedgerAccount) TableName() string { return `"payment-gateway".ledger_accounts` }

type LedgerJournal struct {
	ID            uint64    `gorm:"column:id;primaryKey"`
	ReferenceType string    `gorm:"column:reference_type"`
	ReferenceID   string    `gorm:"column:reference_id"`
	Status        string    `gorm:"column:status"`
	ReversalOfID  *uint64   `gorm:"column:reversal_of_id"`
	Narrative     string    `gorm:"column:narrative"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	PostedAt      time.Time `gorm:"column:posted_at"`
}

func (LedgerJournal) TableName() string { return `"payment-gateway".ledger_journals` }

type LedgerEntry struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	JournalID   uint64    `gorm:"column:journal_id"`
	AccountID   uint64    `gorm:"column:account_id"`
	Side        string    `gorm:"column:side"`
	AmountMinor int64     `gorm:"column:amount_minor"`
	Currency    string    `gorm:"column:currency"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (LedgerEntry) TableName() string { return `"payment-gateway".ledger_entries` }

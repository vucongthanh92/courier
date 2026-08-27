package entities

import "time"

type Wallet struct {
	ID        uint64     `gorm:"column:id;primaryKey;type:bigint"`
	UserID    uint64     `gorm:"column:user_id;type:bigint;not null"`
	Currency  string     `gorm:"column:currency;type:char(3);not null"`
	Status    string     `gorm:"column:status;type:\"payment-gateway\".wallet_status;not null"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	ClosedAt  *time.Time `gorm:"column:closed_at"`
}

func (Wallet) TableName() string { return `"payment-gateway".wallets` }

type WalletBalance struct {
	WalletID       uint64    `gorm:"column:wallet_id;primaryKey;type:bigint"`
	Currency       string    `gorm:"column:currency;type:char(3);not null"`
	AvailableMinor int64     `gorm:"column:available_minor;not null"`
	PendingMinor   int64     `gorm:"column:pending_minor;not null"`
	HeldMinor      int64     `gorm:"column:held_minor;not null"`
	Version        uint64    `gorm:"column:version;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (WalletBalance) TableName() string { return `"payment-gateway".wallet_balances` }

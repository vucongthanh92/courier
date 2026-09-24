package entities

import (
	"fmt"
	"time"
)

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

func (LedgerAccount) TableName() string {
	return `"payment-gateway".ledger_accounts`
}

func (e *LedgerAccount) Initialize(
	accountID uint64,
	accountType string,
	currency string,
	walletID *uint64,
	normalSide string,
) {
	e.ID = accountID
	e.AccountCode = fmt.Sprintf("liability:wallet:%d:vnd", *walletID)
	e.AccountType = accountType
	e.Currency = currency
	e.WalletID = walletID
	e.NormalSide = normalSide
	e.IsActive = true
}

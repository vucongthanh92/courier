package entities

import (
	"time"

	"github.com/vucongthanh92/courier/payment-gateway/helper/constants"
)

type TopUpIntent struct {
	ID                    uint64     `gorm:"column:id;primaryKey"`
	UserID                uint64     `gorm:"column:user_id"`
	WalletID              uint64     `gorm:"column:wallet_id"`
	AmountMinor           int64      `gorm:"column:amount_minor"`
	Currency              string     `gorm:"column:currency"`
	Provider              string     `gorm:"column:provider"`
	Method                string     `gorm:"column:method"`
	Status                string     `gorm:"column:status"`
	ProviderCheckoutID    *string    `gorm:"column:provider_checkout_id"`
	ProviderPaymentURL    *string    `gorm:"column:provider_payment_url"`
	QRPayload             *string    `gorm:"column:qr_payload"`
	ProviderInvoiceNumber string     `gorm:"column:provider_invoice_number"`
	PaymentCode           *string    `gorm:"column:payment_code"`
	ReceivingAccountKey   *string    `gorm:"column:receiving_account_key"`
	ExpiresAt             time.Time  `gorm:"column:expires_at"`
	SucceededAt           *time.Time `gorm:"column:succeeded_at"`
	FailureCode           *string    `gorm:"column:failure_code"`
	FailureMessage        *string    `gorm:"column:failure_message"`
	Metadata              []byte     `gorm:"column:metadata;type:jsonb"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at"`
}

func (TopUpIntent) TableName() string {
	return `"payment-gateway".topup_intents`
}

func (e *TopUpIntent) Initialize(
	topupID uint64,
	userID uint64,
	walletID uint64,
	amountMinor int64,
	currency string,
	provider string,
	method string,
	providerInvoiceNumber string,
	paymentCode string,
) {
	e.ID = topupID
	e.UserID = userID
	e.WalletID = walletID
	e.AmountMinor = amountMinor
	e.Currency = currency
	e.Provider = provider
	e.Method = method
	e.Status = constants.TopupIntentStatusPending
	e.ProviderInvoiceNumber = providerInvoiceNumber
	e.PaymentCode = &paymentCode
	e.ExpiresAt = time.Now().UTC().Add(constants.Time_Cache_15_minutes)
	e.Metadata = []byte(`{}`)
}

func (e *TopUpIntent) CheckAmountMinor(transferAmount int64) bool {
	return e.AmountMinor != transferAmount
}

func (e *TopUpIntent) CheckStatusAndExpires() bool {
	return e.Status != constants.TopupIntentStatusPending || !e.ExpiresAt.After(time.Now().UTC())
}

func (e *TopUpIntent) CheckIsSucceeded() bool {
	return e.Status == constants.TopupIntentStatusSucceeded
}

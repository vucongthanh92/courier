package models

type CreateTopUpRequest struct {
	AmountMinor int64  `json:"amount_minor" validate:"required,gt=0"`
	Method      string `json:"method" validate:"required,oneof=bank_transfer napas_bank_transfer card"`
}

type CheckoutInstruction struct {
	TopUpID        string            `json:"topup_id"`
	InvoiceNumber  string            `json:"invoice_number"`
	ExpiresAt      string            `json:"expires_at"`
	CheckoutAction string            `json:"checkout_action"`
	CheckoutFields map[string]string `json:"checkout_fields"`
}

// SePayBankWebhook is the JSON payload sent by SePay Bank Webhooks.
type SePayBankWebhook struct {
	ID              int64   `json:"id"`
	Gateway         string  `json:"gateway"`
	TransactionDate string  `json:"transactionDate"`
	AccountNumber   string  `json:"accountNumber"`
	SubAccount      *string `json:"subAccount"`
	Code            *string `json:"code"`
	Content         string  `json:"content"`
	TransferType    string  `json:"transferType"`
	Description     string  `json:"description"`
	TransferAmount  int64   `json:"transferAmount"`
	ReferenceCode   string  `json:"referenceCode"`
	Accumulated     int64   `json:"accumulated"`
}

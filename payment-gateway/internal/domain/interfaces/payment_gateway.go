package interfaces

import "context"

type CreateTopUpInput struct {
	InvoiceNumber string
	AmountMinor   int64
	Currency      string
	Method        string
	CustomerID    string
	Description   string
}

type CheckoutInstruction struct {
	Action string
	Fields map[string]string
}

type PaymentGateway interface {
	Name() string
	CreateTopUp(context.Context, CreateTopUpInput) (CheckoutInstruction, error)
}

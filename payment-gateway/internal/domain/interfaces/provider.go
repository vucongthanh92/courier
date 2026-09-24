package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
)

// IdempotencyCommandRepoI defines the interface for managing idempotency keys in the repository.
type ProviderEventCommandRepoI interface {
	CreateIfAbsent(context.Context, *entities.ProviderEvent) (bool, *errHandler.ErrorBuilder)
	MarkIgnored(context.Context, uint64, string) *errHandler.ErrorBuilder
	MarkProcessed(context.Context, uint64) *errHandler.ErrorBuilder
}

// PaymentGateway defines the interface for interacting with different payment gateways.
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

type PaymentGatewayResolverI interface {
	Resolve(context.Context, string) (PaymentGateway, *errHandler.ErrorBuilder)
}

// SePayWebhookServiceI defines the interface for processing SePay bank webhooks.
type SePayWebhookServiceI interface {
	ProcessBankWebhook(context.Context, models.SePayBankWebhook, []byte) (models.WebhookResult, *errHandler.ErrorBuilder)
}

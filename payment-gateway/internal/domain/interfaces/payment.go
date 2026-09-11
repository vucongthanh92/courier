package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
)

type IdempotencyCommandRepoI interface {
	Claim(context.Context, *entities.IdempotencyKey) (*entities.IdempotencyKey, bool, *errHandler.ErrorBuilder)
	Complete(context.Context, uint64, int16, []byte) *errHandler.ErrorBuilder
}
type ProviderEventCommandRepoI interface {
	CreateIfAbsent(context.Context, *entities.ProviderEvent) (bool, *errHandler.ErrorBuilder)
	MarkIgnored(context.Context, uint64, string) *errHandler.ErrorBuilder
	MarkProcessed(context.Context, uint64) *errHandler.ErrorBuilder
}
type LedgerQueryRepoI interface {
	GetWalletLiabilityAccount(context.Context, uint64) (*entities.LedgerAccount, *errHandler.ErrorBuilder)
	GetAccountByCode(context.Context, string) (*entities.LedgerAccount, *errHandler.ErrorBuilder)
}
type LedgerCommandRepoI interface {
	CreateJournalEntries(context.Context, *entities.LedgerJournal, []entities.LedgerEntry) *errHandler.ErrorBuilder
	CreateProviderTransaction(context.Context, *entities.ProviderTransaction) *errHandler.ErrorBuilder
}
type OutboxCommandRepoI interface {
	Create(context.Context, *entities.OutboxEvent) *errHandler.ErrorBuilder
}

type WalletServiceI interface {
	GetOrCreateWallet(context.Context, uint64) (*entities.Wallet, *errHandler.ErrorBuilder)
}

type IdempotencyServiceI interface {
	Claim(context.Context, models.IdempotencyClaimRequest) (*models.IdempotencyClaimResult, *errHandler.ErrorBuilder)
	Complete(context.Context, uint64, int16, any) *errHandler.ErrorBuilder
}

type TopUpServiceI interface {
	CreateTopUp(context.Context, models.CreateTopUpRequest) (models.CheckoutInstruction, bool, *errHandler.ErrorBuilder)
}
type SePayWebhookServiceI interface {
	ProcessBankWebhook(context.Context, models.SePayBankWebhook, []byte) (models.WebhookResult, *errHandler.ErrorBuilder)
}

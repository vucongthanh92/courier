package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
)

type LedgerQueryRepoI interface {
	GetWalletLiabilityAccount(context.Context, uint64) (*entities.LedgerAccount, *errHandler.ErrorBuilder)
	GetAccountByCode(context.Context, string) (*entities.LedgerAccount, *errHandler.ErrorBuilder)
}
type LedgerCommandRepoI interface {
	CreateJournalEntries(context.Context, *entities.LedgerJournal, []entities.LedgerEntry) *errHandler.ErrorBuilder
	CreateProviderTransaction(context.Context, *entities.ProviderTransaction) *errHandler.ErrorBuilder
}

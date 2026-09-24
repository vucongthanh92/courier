package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
)

type WalletQueryRepoI interface {
	GetByUserAndCurrency(context.Context, uint64, string) (*entities.Wallet, *errHandler.ErrorBuilder)
}

type WalletCommandRepoI interface {
	CreateWallet(context.Context, *entities.Wallet, *entities.WalletBalance, *entities.LedgerAccount) *errHandler.ErrorBuilder
	CreditAvailable(context.Context, uint64, int64) *errHandler.ErrorBuilder
}

type WalletServiceI interface {
	GetOrCreateWallet(context.Context, uint64) (*entities.Wallet, *errHandler.ErrorBuilder)
}

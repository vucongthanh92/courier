package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
)

type WalletQueryRepoI interface {
	GetByUserAndCurrency(context.Context, uint64, string) (*entities.Wallet, *errHandler.ErrorBuilder)
	GetBalanceByUserAndCurrency(context.Context, uint64, string) (*entities.WalletBalanceSnapshot, *errHandler.ErrorBuilder)
}

type WalletCommandRepoI interface {
	CreateWallet(context.Context, *entities.Wallet, *entities.WalletBalance, *entities.LedgerAccount) *errHandler.ErrorBuilder
	CreditAvailable(context.Context, uint64, int64) *errHandler.ErrorBuilder
}

type WalletServiceI interface {
	GetOrCreateWallet(context.Context, uint64) (*entities.Wallet, *errHandler.ErrorBuilder)
	GetBalance(context.Context, uint64) (models.WalletBalanceResponse, *errHandler.ErrorBuilder)
}

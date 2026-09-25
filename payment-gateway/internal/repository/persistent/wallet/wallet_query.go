package wallet

import (
	"context"
	"github.com/vucongthanh92/courier/payment-gateway/database"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/transaction"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"gorm.io/gorm"
)

type walletQueryRepository struct {
	readDB *gorm.DB
}

func InitWalletQueryRepository(db *database.GormReadDb) interfaces.WalletQueryRepoI {
	return &walletQueryRepository{readDB: *db}
}
func (r *walletQueryRepository) GetByUserAndCurrency(ctx context.Context, userID uint64, currency string) (*entities.Wallet, *errHandler.ErrorBuilder) {
	var e entities.Wallet
	err := transaction.RunnerFromCtx(ctx, r.readDB).Where("user_id = ? AND currency = ?", userID, currency).First(&e).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &e, nil
}

func (r *walletQueryRepository) GetBalanceByUserAndCurrency(ctx context.Context, userID uint64, currency string) (*entities.WalletBalanceSnapshot, *errHandler.ErrorBuilder) {
	var snapshot entities.WalletBalanceSnapshot

	err := transaction.RunnerFromCtx(ctx, r.readDB).
		Table(`"payment-gateway".wallets AS wallets`).
		Select(`wallets.id AS wallet_id, wallets.currency, wallets.status, balances.available_minor, balances.pending_minor, balances.held_minor, balances.updated_at`).
		Joins(`JOIN "payment-gateway".wallet_balances AS balances ON balances.wallet_id = wallets.id`).
		Where("wallets.user_id = ? AND wallets.currency = ?", userID, currency).
		Take(&snapshot).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	return &snapshot, nil
}

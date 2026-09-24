package ledger

import (
	"context"
	"github.com/vucongthanh92/courier/payment-gateway/database"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/transaction"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"gorm.io/gorm"
)

type ledgerQueryRepository struct {
	readDB *gorm.DB
}

func InitLedgerQueryRepository(db *database.GormReadDb) interfaces.LedgerQueryRepoI {
	return &ledgerQueryRepository{readDB: *db}
}
func (r *ledgerQueryRepository) GetWalletLiabilityAccount(ctx context.Context, walletID uint64) (*entities.LedgerAccount, *errHandler.ErrorBuilder) {
	var e entities.LedgerAccount
	err := transaction.RunnerFromCtx(ctx, r.readDB).Where("wallet_id = ? AND account_type = ?", walletID, "liability").First(&e).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &e, nil
}
func (r *ledgerQueryRepository) GetAccountByCode(ctx context.Context, code string) (*entities.LedgerAccount, *errHandler.ErrorBuilder) {
	var e entities.LedgerAccount
	err := transaction.RunnerFromCtx(ctx, r.readDB).Where("account_code = ?", code).First(&e).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &e, nil
}

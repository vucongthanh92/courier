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

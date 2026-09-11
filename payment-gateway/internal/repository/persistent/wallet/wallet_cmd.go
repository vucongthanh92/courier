package wallet

import (
	"context"
	"github.com/vucongthanh92/courier/payment-gateway/database"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/transaction"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"gorm.io/gorm"
	"time"
)

type walletCmdRepository struct {
	writeDB *gorm.DB
}

func InitWalletCmdRepository(db *database.GormWriteDb) interfaces.WalletCommandRepoI {
	return &walletCmdRepository{writeDB: *db}
}
func (r *walletCmdRepository) CreateWallet(ctx context.Context, wallet *entities.Wallet, balance *entities.WalletBalance, account *entities.LedgerAccount) *errHandler.ErrorBuilder {
	run := transaction.RunnerFromCtx(ctx, r.writeDB)
	if err := run.Create(wallet).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	if err := run.Create(balance).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	if err := run.Create(account).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}
func (r *walletCmdRepository) CreditAvailable(ctx context.Context, walletID uint64, amount int64) *errHandler.ErrorBuilder {
	run := transaction.RunnerFromCtx(ctx, r.writeDB)
	result := run.Model(&entities.WalletBalance{}).Where("wallet_id = ?", walletID).Updates(map[string]any{"available_minor": gorm.Expr("available_minor + ?", amount), "version": gorm.Expr("version + 1"), "updated_at": time.Now().UTC()})
	if result.Error != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(result.Error)
	}
	if result.RowsAffected != 1 {
		return errHandler.InitErrorBuilder(ctx).ValidateError(gorm.ErrRecordNotFound)
	}
	return nil
}

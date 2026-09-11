package topup

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

type topUpCmdRepository struct {
	writeDB *gorm.DB
}

func InitTopUpCmdRepository(db *database.GormWriteDb) interfaces.TopUpCommandRepoI {
	return &topUpCmdRepository{writeDB: *db}
}
func (r *topUpCmdRepository) Create(ctx context.Context, e *entities.TopUpIntent) *errHandler.ErrorBuilder {
	if err := transaction.RunnerFromCtx(ctx, r.writeDB).Create(e).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}
func (r *topUpCmdRepository) MarkSucceeded(ctx context.Context, id uint64, account string) *errHandler.ErrorBuilder {
	err := transaction.RunnerFromCtx(ctx, r.writeDB).Model(&entities.TopUpIntent{}).Where("id = ?", id).Updates(map[string]any{"status": "succeeded", "succeeded_at": time.Now().UTC(), "receiving_account_key": account, "updated_at": time.Now().UTC()}).Error
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

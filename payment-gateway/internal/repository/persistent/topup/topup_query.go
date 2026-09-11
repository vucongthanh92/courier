package topup

import (
	"context"
	"github.com/vucongthanh92/courier/payment-gateway/database"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/transaction"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type topUpQueryRepository struct {
	readDB *gorm.DB
}

func InitTopUpQueryRepository(db *database.GormReadDb) interfaces.TopUpQueryRepoI {
	return &topUpQueryRepository{readDB: *db}
}
func (r *topUpQueryRepository) GetForUpdateByPaymentCode(ctx context.Context, provider, code string) (*entities.TopUpIntent, *errHandler.ErrorBuilder) {
	var e entities.TopUpIntent
	run := transaction.RunnerFromCtx(ctx, r.readDB)
	err := run.Clauses(clause.Locking{Strength: "UPDATE"}).Where("provider = ? AND payment_code = ?", provider, code).First(&e).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &e, nil
}

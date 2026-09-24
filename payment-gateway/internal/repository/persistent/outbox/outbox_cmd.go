package outbox

import (
	"context"
	"github.com/vucongthanh92/courier/payment-gateway/database"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/transaction"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"gorm.io/gorm"
)

type outboxCmdRepository struct {
	writeDB *gorm.DB
}

func InitOutboxCmdRepository(db *database.GormWriteDb) interfaces.OutboxCommandRepoI {
	return &outboxCmdRepository{writeDB: *db}
}
func (r *outboxCmdRepository) Create(ctx context.Context, e *entities.OutboxEvent) *errHandler.ErrorBuilder {
	if err := transaction.RunnerFromCtx(ctx, r.writeDB).Create(e).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

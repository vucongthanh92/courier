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

type ledgerCmdRepository struct {
	writeDB *gorm.DB
}

func InitLedgerCmdRepository(db *database.GormWriteDb) interfaces.LedgerCommandRepoI {
	return &ledgerCmdRepository{writeDB: *db}
}
func (r *ledgerCmdRepository) CreateJournalEntries(ctx context.Context, j *entities.LedgerJournal, e []entities.LedgerEntry) *errHandler.ErrorBuilder {
	run := transaction.RunnerFromCtx(ctx, r.writeDB)
	if err := run.Create(j).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	if err := run.Create(&e).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}
func (r *ledgerCmdRepository) CreateProviderTransaction(ctx context.Context, e *entities.ProviderTransaction) *errHandler.ErrorBuilder {
	if err := transaction.RunnerFromCtx(ctx, r.writeDB).Create(e).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

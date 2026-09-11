package provider_event

import (
	"context"
	"github.com/vucongthanh92/courier/payment-gateway/database"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/helper/transaction"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type providerEventCmdRepository struct {
	writeDB *gorm.DB
}

func InitProviderEventCmdRepository(db *database.GormWriteDb) interfaces.ProviderEventCommandRepoI {
	return &providerEventCmdRepository{writeDB: *db}
}
func (r *providerEventCmdRepository) CreateIfAbsent(ctx context.Context, e *entities.ProviderEvent) (bool, *errHandler.ErrorBuilder) {
	result := transaction.RunnerFromCtx(ctx, r.writeDB).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "provider"}, {Name: "provider_event_id"}}, DoNothing: true}).Create(e)
	if result.Error != nil {
		return false, errHandler.InitErrorBuilder(ctx).ValidateError(result.Error)
	}
	return result.RowsAffected == 1, nil
}
func (r *providerEventCmdRepository) MarkIgnored(ctx context.Context, id uint64, code string) *errHandler.ErrorBuilder {
	return r.update(ctx, id, map[string]any{"status": "ignored", "error_code": code, "processed_at": time.Now().UTC()})
}
func (r *providerEventCmdRepository) MarkProcessed(ctx context.Context, id uint64) *errHandler.ErrorBuilder {
	return r.update(ctx, id, map[string]any{"status": "processed", "processed_at": time.Now().UTC()})
}
func (r *providerEventCmdRepository) update(ctx context.Context, id uint64, updates map[string]any) *errHandler.ErrorBuilder {
	err := transaction.RunnerFromCtx(ctx, r.writeDB).Model(&entities.ProviderEvent{}).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

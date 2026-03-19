package refreshtoken

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/user-service/database"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"
)

type refreshTokenCmdRepo struct {
	writeDb *gorm.DB
}

func InitRefreshTokenCmdRepository(writeDb *database.GormWriteDb) interfaces.RefreshTokenCommandRepoI {
	return &refreshTokenCmdRepo{writeDb: *writeDb}
}

func (r *refreshTokenCmdRepo) Insert(ctx context.Context, entity *entities.RefreshToken) *errHandler.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "InsertRefreshToken")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, r.writeDb)
	if err := run.Model(&entities.RefreshToken{}).Create(entity).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

func (r *refreshTokenCmdRepo) RevokeByID(ctx context.Context, id uint64, revokedAt time.Time) *errHandler.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "RevokeRefreshToken")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, r.writeDb)
	err := run.Model(&entities.RefreshToken{}).Where("id = ?", id).Update("revoked_at", revokedAt).Error
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

func (r *refreshTokenCmdRepo) RevokeByUser(ctx context.Context, userID uint64, revokedAt time.Time) *errHandler.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "RevokeRefreshTokenByUser")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, r.writeDb)
	err := run.Model(&entities.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", revokedAt).Error
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

func (r *refreshTokenCmdRepo) Rotate(ctx context.Context, oldID uint64, newEntity *entities.RefreshToken) (*entities.RefreshToken, *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "RotateRefreshToken")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, r.writeDb)
	if err := run.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entities.RefreshToken{}).
			Where("id = ?", oldID).
			Updates(map[string]interface{}{"revoked_at": newEntity.CreatedAt, "replaced_by_id": newEntity.ID}).Error; err != nil {
			return err
		}
		return tx.Model(&entities.RefreshToken{}).Create(newEntity).Error
	}); err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return newEntity, nil
}

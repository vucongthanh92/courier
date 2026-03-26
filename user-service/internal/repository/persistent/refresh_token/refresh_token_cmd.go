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

// UpsertByUserAgent will upsert refresh token by user_id and user_agent,
// if exist then update expires_at, otherwise insert new record
func (r *refreshTokenCmdRepo) UpsertByUserAgent(ctx context.Context, entity *entities.RefreshToken) *errHandler.ErrorBuilder {

	// Start tracing span for upsert operation
	ctx, span := tracing.StartSpanFromContext(ctx, "UpsertRefreshTokenByUserAgent")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, r.writeDb)

	// Get user agent from entity, if nil then use empty string to prevent null value in database which can cause issues with unique index
	ua := "unknown"
	if entity.UserAgent != nil {
		ua = *entity.UserAgent
	}

	// Try to update existing record first
	res := run.Model(&entities.RefreshToken{}).
		Where("user_id = ? AND user_agent = ? AND revoked_at IS NULL", entity.UserID, ua).
		Updates(map[string]interface{}{
			"token_hash": entity.TokenHash,
			"expires_at": entity.ExpiresAt,
		})
	if res.Error != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(res.Error)
	}

	// If existing record found and updated,
	// return without inserting new record to prevent multiple active refresh tokens for same user and device
	if res.RowsAffected > 0 {
		return nil
	}

	// No existing record, insert new one
	if err := run.Model(&entities.RefreshToken{}).Create(entity).Error; err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	// Inserted new record successfully
	return nil
}

// RevokeByID will set revoked_at by id to revoke refresh token
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

// RevokeByUser will set revoked_at by user_id to revoke all refresh tokens of user (optional for logout-all)
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

// Rotate will revoke old refresh token and insert new refresh token in one transaction to ensure atomicity
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

func (r *refreshTokenCmdRepo) DeleteExpiredAndRevoked(ctx context.Context, now time.Time) error {
	run := transaction.RunnerFromCtx(ctx, r.writeDb)
	return run.
		Where("(revoked_at IS NOT NULL) OR (expires_at <= ?)", now).
		Delete(&entities.RefreshToken{}).Error
}

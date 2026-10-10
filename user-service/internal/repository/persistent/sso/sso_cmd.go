package sso

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
	"github.com/vucongthanh92/go-base-utils/helper/transaction"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"
)

type ssoSessionCommandRepo struct {
	writeDb *gorm.DB
}

func InitSsoSessionCommandRepo(writeDb *database.GormWriteDb) interfaces.SsoSessionCommandRepoI {
	return &ssoSessionCommandRepo{writeDb: *writeDb}
}

func (r *ssoSessionCommandRepo) CreateSession(ctx context.Context, entity *entities.SsoSession) *utilsError.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "CreateSsoSession")
	defer span.End()

	run := transaction.RunnerFromCtx(ctx, r.writeDb)
	if err := run.Model(&entities.SsoSession{}).Create(entity).Error; err != nil {
		return utilsError.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

func (r *ssoSessionCommandRepo) RevokeByHash(ctx context.Context, sessionHash string, revokedAt time.Time) *utilsError.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "RevokeSsoSessionByHash")
	defer span.End()

	run := transaction.RunnerFromCtx(ctx, r.writeDb)
	err := run.Model(&entities.SsoSession{}).
		Where("session_hash = ? AND revoked_at IS NULL", sessionHash).
		Update("revoked_at", revokedAt).Error
	if err != nil {
		return utilsError.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

func (r *ssoSessionCommandRepo) RevokeByUser(ctx context.Context, userID uint64, revokedAt time.Time) *utilsError.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "RevokeSsoSessionByUser")
	defer span.End()

	run := transaction.RunnerFromCtx(ctx, r.writeDb)
	err := run.Model(&entities.SsoSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", revokedAt).Error
	if err != nil {
		return utilsError.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

type ssoAuthorizationCodeCommandRepo struct {
	writeDb *gorm.DB
}

func InitSsoAuthorizationCodeCommandRepo(writeDb *database.GormWriteDb) interfaces.SsoAuthorizationCodeCommandRepoI {
	return &ssoAuthorizationCodeCommandRepo{writeDb: *writeDb}
}

func (r *ssoAuthorizationCodeCommandRepo) CreateCode(ctx context.Context, entity *entities.SsoAuthorizationCode) *utilsError.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "CreateSsoAuthorizationCode")
	defer span.End()

	run := transaction.RunnerFromCtx(ctx, r.writeDb)
	if err := run.Model(&entities.SsoAuthorizationCode{}).Create(entity).Error; err != nil {
		return utilsError.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

func (r *ssoAuthorizationCodeCommandRepo) ConsumeCode(ctx context.Context, codeHash string, consumedAt time.Time) *utilsError.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "ConsumeSsoAuthorizationCode")
	defer span.End()

	run := transaction.RunnerFromCtx(ctx, r.writeDb)
	err := run.Model(&entities.SsoAuthorizationCode{}).
		Where("code_hash = ? AND consumed_at IS NULL", codeHash).
		Update("consumed_at", consumedAt).Error
	if err != nil {
		return utilsError.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

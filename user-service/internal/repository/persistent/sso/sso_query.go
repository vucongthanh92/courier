package sso

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

type ssoSessionQueryRepo struct {
	readDb *gorm.DB
}

func InitSsoSessionQueryRepo(readDb *database.GormReadDb) interfaces.SsoSessionQueryRepoI {
	return &ssoSessionQueryRepo{readDb: *readDb}
}

func (r *ssoSessionQueryRepo) GetActiveByHash(
	ctx context.Context,
	sessionHash string,
	now time.Time,
) (*entities.SsoSession, *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "GetActiveSsoSessionByHash")
	defer span.End()

	var session entities.SsoSession
	run := transaction.RunnerFromCtx(ctx, r.readDb)
	err := run.Model(&entities.SsoSession{}).
		Where("session_hash = ? AND revoked_at IS NULL AND expires_at > ?", sessionHash, now).
		Take(&session).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &session, nil
}

type ssoAuthorizationCodeQueryRepo struct {
	readDb *gorm.DB
}

func InitSsoAuthorizationCodeQueryRepo(readDb *database.GormReadDb) interfaces.SsoAuthorizationCodeQueryRepoI {
	return &ssoAuthorizationCodeQueryRepo{readDb: *readDb}
}

func (r *ssoAuthorizationCodeQueryRepo) GetActiveByHash(
	ctx context.Context,
	codeHash string,
	now time.Time,
) (*entities.SsoAuthorizationCode, *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "GetActiveSsoAuthorizationCodeByHash")
	defer span.End()

	var code entities.SsoAuthorizationCode
	run := transaction.RunnerFromCtx(ctx, r.readDb)
	err := run.Model(&entities.SsoAuthorizationCode{}).
		Where("code_hash = ? AND consumed_at IS NULL AND expires_at > ?", codeHash, now).
		Take(&code).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return &code, nil
}

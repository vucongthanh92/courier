package jwk

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"
)

type jwkQueryRepo struct {
	readDb *gorm.DB
}

func InitJWKQueryRepository(readDb *database.GormReadDb) interfaces.JWKQueryRepoI {
	return &jwkQueryRepo{readDb: *readDb}
}

// GetActiveKey implements interfaces.JWKQueryRepoI
func (r *jwkQueryRepo) GetActiveKey(ctx context.Context) (entities.JWKKey, *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "GetActiveJWKKey")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, r.readDb)

	var res entities.JWKKey
	err := run.Model(&entities.JWKKey{}).Where("active = true").Order("created_at DESC").Take(&res).Error
	if err != nil {
		return res, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	return res, nil
}

// GetKeyByKid implements interfaces.JWKQueryRepoI.
func (r *jwkQueryRepo) GetKeyByKid(ctx context.Context, kid string) (entities.JWKKey, *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "GetJWKKeyByKid")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, r.readDb)

	var res entities.JWKKey
	err := run.Model(&entities.JWKKey{}).
		Where("kid = ?", kid).
		Take(&res).Error
	if err != nil {
		return res, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	return res, nil
}

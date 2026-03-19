package authcredential

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type authCredQueryRepo struct {
	readDb *gorm.DB
}

func InitAuthCredentialQueryRepository(readDb *database.GormReadDb) interfaces.AuthCredentialQueryRepoI {
	return &authCredQueryRepo{readDb: *readDb}
}

func (repo *authCredQueryRepo) GetByUserID(ctx context.Context, userID uint64) (entities.AuthCredential, *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "GetAuthCredentialByUserID")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.readDb)

	var res entities.AuthCredential
	err := run.Model(&entities.AuthCredential{}).Where("user_id = ?", userID).Take(&res).Error
	if err != nil {
		return res, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return res, nil
}

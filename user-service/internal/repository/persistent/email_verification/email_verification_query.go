package emailverification

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

type emailVerificationQueryRepository struct {
	readDb *gorm.DB
}

func InitEmailVerificationQueryRepository(readDb *database.GormReadDb) interfaces.EmailVerificationQueryRepoI {
	return &emailVerificationQueryRepository{readDb: *readDb}
}

func (repo *emailVerificationQueryRepository) GetActiveByEmail(ctx context.Context, email string) (entities.EmailVerification, *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "GetActiveEmailVerification")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.readDb)

	var res entities.EmailVerification
	err := run.Where("email = ? AND used_at IS NULL AND expires_at > now()", email).
		Order("created_at DESC").
		Take(&res).Error
	if err != nil {
		return res, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return res, nil
}

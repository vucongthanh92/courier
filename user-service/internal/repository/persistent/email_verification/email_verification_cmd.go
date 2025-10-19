package emailverification

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type emailVerificationCmdRepository struct {
	writeDb *gorm.DB
}

func InitEmailVerificationCmdRepository(writeDb *database.GormWriteDb) interfaces.EmailVerificationCommandRepoI {
	return &emailVerificationCmdRepository{
		writeDb: *writeDb,
	}
}

func (repo *emailVerificationCmdRepository) InsertEmailVerification(ctx context.Context, entity entities.EmailVerification) (
	entities.EmailVerification, *errHandler.ErrorBuilder) {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "InsertEmailVerification")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	// Insert email verification record
	err := run.Model(entities.EmailVerification{}).Create(&entity).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return entity, resErr
	}

	return entity, nil
}

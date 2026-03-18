package emailverification

import (
	"context"
	"time"

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

// InsertEmailVerification implements interfaces.EmailVerificationCommandRepoI
// inserts a new email verification record into the database.
func (repo *emailVerificationCmdRepository) InsertEmailVerification(ctx context.Context, entity *entities.EmailVerification) *errHandler.ErrorBuilder {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "InsertEmailVerification")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	// Insert email verification record
	err := run.Model(entities.EmailVerification{}).Create(entity).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return resErr
	}

	return nil
}

// UpdateToken implements interfaces.EmailVerificationCommandRepoI
// updates the token hash and expiration time for an active email verification record associated with the given email.
func (repo *emailVerificationCmdRepository) UpdateToken(ctx context.Context, email, tokenHash string, expiresAt time.Time) *errHandler.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "UpdateTokenEmailVerification")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	err := run.Model(&entities.EmailVerification{}).
		Where("email = ? AND used_at IS NULL", email).
		Updates(map[string]interface{}{
			"token_hash": tokenHash,
			"expires_at": expiresAt,
		}).Error
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

// MarkUsed implements interfaces.EmailVerificationCommandRepoI
// marks an email verification record as used by setting the used_at timestamp.
func (repo *emailVerificationCmdRepository) MarkUsed(ctx context.Context, id uint64, usedAt time.Time) *errHandler.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "MarkUsedEmailVerification")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	err := run.Model(&entities.EmailVerification{}).
		Where("id = ?", id).
		Update("used_at", usedAt).Error
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

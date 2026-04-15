package authcredential

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

type authCredentialCmdRepository struct {
	writeDb *gorm.DB
}

func InitAuthCredentialCmdRepository(writeDb *database.GormWriteDb) interfaces.AuthCredentialCommandRepoI {
	return &authCredentialCmdRepository{
		writeDb: *writeDb,
	}
}

// InsertAuthCredential inserts a new auth credential record into the database.
func (repo *authCredentialCmdRepository) InsertAuthCredential(ctx context.Context, entity *entities.AuthCredential) *errHandler.ErrorBuilder {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "InsertAuthCredential")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	// Insert auth credential record
	err := run.Model(entities.AuthCredential{}).Create(entity).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return resErr
	}

	return nil
}

// UpdatePassword updates the password fields of the auth credential record for the specified user ID.
func (repo *authCredentialCmdRepository) UpdatePassword(ctx context.Context, req *entities.AuthCredential) *errHandler.ErrorBuilder {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "UpdatePassword")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDb)

	// Update password fields
	err := run.Model(&entities.AuthCredential{}).
		Where("user_id = ?", req.UserID).
		Updates(map[string]interface{}{
			"password_hash":    req.PasswordHash,
			"password_algo":    req.PasswordAlgo,
			"password_version": req.PasswordVersion,
		}).Error
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	return nil
}

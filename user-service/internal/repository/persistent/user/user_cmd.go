package user

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type userCmdRepository struct {
	writeDB *gorm.DB
}

func InitUserCmdRepository(writeDB *database.GormWriteDb) interfaces.UserCommandRepoI {
	return &userCmdRepository{
		writeDB: *writeDB,
	}
}

// InsertUser implements interfaces.UserCommandRepoI
// This method inserts a new user record into the database.
// It takes a user entity as a parameter and returns an error builder if any error occurs during the insertion process.
func (repo *userCmdRepository) InsertUser(ctx context.Context, entity *entities.User) *errHandler.ErrorBuilder {

	// Start tracing span
	ctx, span := tracing.StartSpanFromContext(ctx, "InsertUser")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDB)

	// Insert user record
	err := run.Model(entities.User{}).Create(entity).Error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return resErr
	}

	return nil
}

// UpdateEmailVerified implements interfaces.UserCommandRepoI
// This method updates the email verification status of a user. It takes the user ID and the new status as parameters.
func (repo *userCmdRepository) UpdateEmailVerified(ctx context.Context, id uint64, status string) *errHandler.ErrorBuilder {
	ctx, span := tracing.StartSpanFromContext(ctx, "UpdateEmailVerified")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.writeDB)

	err := run.Model(&entities.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         status,
			"email_verified": status == "verified",
		}).Error
	if err != nil {
		return errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return nil
}

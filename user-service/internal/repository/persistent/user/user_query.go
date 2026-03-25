package user

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

type userQueryRepository struct {
	readDb *gorm.DB
}

func InitUserQueryRepository(readDb *database.GormReadDb) interfaces.UserQueryRepoI {
	return &userQueryRepository{
		readDb: *readDb,
	}
}

// GetUserByID implements interfaces.UserQueryRepoI
// This method retrieves a user record from the database based on the provided user ID.
// It returns the user entity and an error builder if any error occurs during the retrieval process.
func (repo *userQueryRepository) GetUserByID(ctx context.Context, id uint64) (
	res entities.User, errRes *errHandler.ErrorBuilder) {

	// Start tracing
	ctx, span := tracing.StartSpanFromContext(ctx, "GetUserByID")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.readDb)

	// Query user by ID
	err := run.Model(&entities.User{}).Select("*").
		Where("id = ?", id).Where("deleted_at is null").
		Take(&res).Error

	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return res, resErr
	}

	return res, errRes
}

// CheckExistingEmailOrPhone implements interfaces.UserQueryRepoI
// This method checks if a user with the given email or phone number already exists in the database.
// It returns a boolean indicating existence and an error builder if any error occurs during the check.
func (repo *userQueryRepository) CheckExistingEmailOrPhone(ctx context.Context, email string, phoneNumber string) (
	res bool, errRes *errHandler.ErrorBuilder) {

	// Start tracing
	ctx, span := tracing.StartSpanFromContext(ctx, "CheckExistingEmailOrPhone")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.readDb)

	// Query existing email or phone number
	err := run.Raw(`SELECT EXISTS (
            SELECT 1 FROM "user-service".users WHERE email = ? OR phone_number = ?
        )`, email, phoneNumber).Scan(&res).Error

	// Handle error
	if err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return res, resErr
	}

	return res, nil
}

// GetUserByEmail implements interfaces.UserQueryRepoI
// This method retrieves a user record from the database based on the provided email.
// It returns the user entity and an error builder if any error occurs during the retrieval process.
func (repo *userQueryRepository) GetUserByEmail(ctx context.Context, email string) (entities.User, *errHandler.ErrorBuilder) {
	ctx, span := tracing.StartSpanFromContext(ctx, "GetUserByEmail")
	defer span.End()
	run := transaction.RunnerFromCtx(ctx, repo.readDb)

	var res entities.User
	err := run.Model(&entities.User{}).Where("email = ? AND deleted_at IS NULL", email).Take(&res).Error
	if err != nil {
		return res, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}
	return res, nil
}

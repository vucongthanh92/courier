package user

import (
	"context"
	"strings"

	"github.com/vucongthanh92/courier/user-service/database"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/helper/transaction"
	"github.com/vucongthanh92/go-base-utils/tracing"
	"gorm.io/gorm"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type userQueryRepository struct {
	readDb *gorm.DB
}

func InitUserQueryRepository(readDb *database.GormReadDb) interfaces.UserQueryRepoI {
	return &userQueryRepository{
		readDb: *readDb,
	}
}

// GetUserByIdOrEmail implements interfaces.UserQueryRepoI
// This method retrieves a user record from the database based on the provided user ID.
// It returns the user entity and an error builder if any error occurs during the retrieval process.
func (repo *userQueryRepository) GetUserByIdOrEmail(ctx context.Context, req models.GetUserByIdOrEmailRequest) (
	res *entities.User, errRes *errHandler.ErrorBuilder) {

	// Start tracing
	ctx, span := tracing.StartSpanFromContext(ctx, "GetUserByIdOrEmail")
	defer span.End()
	runner := transaction.RunnerFromCtx(ctx, repo.readDb)

	// Build query
	ors := []string{}
	args := []interface{}{}
	runner = runner.Model(&entities.User{}).Select("*").Where("deleted_at is null")

	// Add conditions based on provided parameters
	if req.UserID != nil {
		ors = append(ors, "id = ?")
		args = append(args, *req.UserID)
	}

	// If email is provided, add OR condition for email
	if req.Email != nil {
		ors = append(ors, "email = ?")
		args = append(args, *req.Email)
	}

	// Combine OR conditions
	if len(ors) > 0 {
		runner = runner.Where("("+strings.Join(ors, " OR ")+")", args...)
	}

	// Execute query and handle result
	var user entities.User
	if err := runner.Take(&user).Error; err != nil {
		resErr := errHandler.InitErrorBuilder(ctx).ValidateError(err)
		return nil, resErr
	}

	return &user, nil
}

// GetUsersByIDs returns users for the provided IDs, excluding soft-deleted rows.
func (repo *userQueryRepository) GetUsersByIDs(ctx context.Context, userIDs []uint64) (
	res []entities.User, errRes *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "GetUsersByIDs")
	defer span.End()

	if len(userIDs) == 0 {
		return []entities.User{}, nil
	}

	runner := transaction.RunnerFromCtx(ctx, repo.readDb)
	err := runner.Model(&entities.User{}).
		Select("id, display_name, avatar_url, status").
		Where("deleted_at is null").
		Where("id IN ?", userIDs).
		Find(&res).Error
	if err != nil {
		return nil, errHandler.InitErrorBuilder(ctx).ValidateError(err)
	}

	return res, nil
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

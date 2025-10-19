package category

import (
	"context"

	"github.com/jinzhu/copier"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/tracing"
)

type UserServiceImpl struct {
	userReadRepo  interfaces.UserQueryRepoI
	userWriteRepo interfaces.UserCommandRepoI
}

func InitUserService(
	userReadRepo interfaces.UserQueryRepoI,
	userWriteRepo interfaces.UserCommandRepoI,
) interfaces.UserServiceI {
	return &UserServiceImpl{
		userReadRepo:  userReadRepo,
		userWriteRepo: userWriteRepo,
	}
}

// Signup implements interfaces.UserServiceI
func (s *UserServiceImpl) Signup(ctx context.Context, req models.SignupRequest) (
	entities.User, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "Signup")
	defer span.End()

	// step 1. Map request to entity
	userEntity := entities.User{}
	copier.Copy(&userEntity, &req)

	// step 2. check email and phone number exist with user existing

	// step 3. init transaction to create user with
	// table users, email_verification, auth_credentials, audit_log, ...
	res, resErr := s.userWriteRepo.InsertUser(ctx, userEntity)
	if resErr != nil {
		return res, resErr
	}

	// step 4. handle after created user successfully, send verify email, sms, ...

	return res, nil
}

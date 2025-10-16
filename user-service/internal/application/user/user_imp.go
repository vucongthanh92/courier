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

func (s *UserServiceImpl) CreateUser(ctx context.Context, req models.CreateUserRequest) (
	entities.User, *errHandler.ErrorBuilder) {

	ctx, span := tracing.StartSpanFromContext(ctx, "CreateUser")
	defer span.End()

	userEntity := entities.User{}
	copier.Copy(&userEntity, &req)

	res, resErr := s.userWriteRepo.InsertUser(ctx, userEntity)
	if resErr != nil {
		return res, resErr
	}

	return res, nil
}

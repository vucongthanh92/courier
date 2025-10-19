package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

// repository interface
type UserQueryRepoI interface {
}

type UserCommandRepoI interface {
	InsertUser(ctx context.Context, entity entities.User) (
		entities.User, *errHandler.ErrorBuilder)
}

// service interface
type UserServiceI interface {
	Signup(ctx context.Context, req models.SignupRequest) (
		*entities.User, *errHandler.ErrorBuilder)
}

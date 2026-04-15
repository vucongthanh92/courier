package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type AuthCredentialQueryRepoI interface {
	GetByUserID(ctx context.Context, userID uint64) (entities.AuthCredential, *errHandler.ErrorBuilder)
}

type AuthCredentialCommandRepoI interface {
	InsertAuthCredential(ctx context.Context, entity *entities.AuthCredential) *errHandler.ErrorBuilder
	UpdatePassword(ctx context.Context, req *entities.AuthCredential) *errHandler.ErrorBuilder
}

type AuthCredentialServiceI interface {
	SetPassword(ctx context.Context, req models.GeneratePasswordRequest) *errHandler.ErrorBuilder
}

package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type AuthCredentialQueryRepoI interface {
}

type AuthCredentialCommandRepoI interface {
	InsertAuthCredential(ctx context.Context, entity *entities.AuthCredential) *errHandler.ErrorBuilder
}

type AuthCredentialServiceI interface {
}

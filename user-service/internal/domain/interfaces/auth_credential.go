package interfaces

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
)

type AuthCredentialQueryRepoI interface {
	GetByUserID(ctx context.Context, userID uint64) (entities.AuthCredential, *utilsError.ErrorBuilder)
}

type AuthCredentialCommandRepoI interface {
	InsertAuthCredential(ctx context.Context, entity *entities.AuthCredential) *utilsError.ErrorBuilder
	UpdatePassword(ctx context.Context, req *entities.AuthCredential) *utilsError.ErrorBuilder
}

type AuthCredentialServiceI interface {
	SetPassword(ctx context.Context, req models.GeneratePasswordRequest) *utilsError.ErrorBuilder
}

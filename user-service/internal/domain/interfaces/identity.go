package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type IdentityQueryRepoI interface {
}

type IdentityCommandRepoI interface {
	InserIdentity(ctx context.Context, entity entities.Identity) (
		entities.Identity, *errHandler.ErrorBuilder)
}

type IdentityServiceI interface {
	CreateIdentity(ctx context.Context, req models.CreateIdentityParams) (
		entities.Identity, *errHandler.ErrorBuilder)
}

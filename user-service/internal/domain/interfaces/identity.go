package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type IdentityQueryRepoI interface {
	GetByProviderUID(ctx context.Context, provider, providerUID string) (*entities.Identity, *errHandler.ErrorBuilder)
}

type IdentityCommandRepoI interface {
	InserIdentity(ctx context.Context, entity *entities.Identity) *errHandler.ErrorBuilder
}

type IdentityServiceI interface {
	OAuthLogin(ctx context.Context, req models.OAuthLoginRequest) (*models.OAuthLoginResponse, *errHandler.ErrorBuilder)
	OAuthCallback(ctx context.Context, req models.OAuthCallbackRequest) (*models.OAuthLoginResponse, *errHandler.ErrorBuilder)
}

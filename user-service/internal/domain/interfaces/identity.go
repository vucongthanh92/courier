package interfaces

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
)

type IdentityQueryRepoI interface {
	GetByProviderUID(ctx context.Context, provider, providerUID string) (*entities.Identity, *utilsError.ErrorBuilder)
}

type IdentityCommandRepoI interface {
	InserIdentity(ctx context.Context, entity *entities.Identity) *utilsError.ErrorBuilder
}

type IdentityUseCaseI interface {
	OAuthLogin(ctx context.Context, req models.OAuthLoginRequest) (*models.JwtTokenResponse, *utilsError.ErrorBuilder)
	OAuthCallback(ctx context.Context, req models.OAuthCallbackRequest) (*models.JwtTokenResponse, *utilsError.ErrorBuilder)
}

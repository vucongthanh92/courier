package interfaces

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type ProviderClient interface {
	Verify(ctx context.Context, token string) (models.ProviderProfile, error)
}

type GoogleProviderClient interface {
	Verify(ctx context.Context, token string) (models.ProviderProfile, error)
	ExchangeCode(ctx context.Context, code, redirectURI string) (string, error)
}

type GithubProviderClient interface {
	Verify(ctx context.Context, token string) (models.ProviderProfile, error)
}

type GithubCodeExchanger interface {
	ExchangeCode(ctx context.Context, code, redirectURI string) (accessToken string, err error)
}

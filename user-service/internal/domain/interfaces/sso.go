package interfaces

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
)

type SsoSessionCommandRepoI interface {
	CreateSession(ctx context.Context, entity *entities.SsoSession) *utilsError.ErrorBuilder
	RevokeByHash(ctx context.Context, sessionHash string, revokedAt time.Time) *utilsError.ErrorBuilder
	RevokeByUser(ctx context.Context, userID uint64, revokedAt time.Time) *utilsError.ErrorBuilder
}

type SsoSessionQueryRepoI interface {
	GetActiveByHash(ctx context.Context, sessionHash string, now time.Time) (*entities.SsoSession, *utilsError.ErrorBuilder)
}

type SsoAuthorizationCodeCommandRepoI interface {
	CreateCode(ctx context.Context, entity *entities.SsoAuthorizationCode) *utilsError.ErrorBuilder
	ConsumeCode(ctx context.Context, codeHash string, consumedAt time.Time) *utilsError.ErrorBuilder
}

type SsoAuthorizationCodeQueryRepoI interface {
	GetActiveByHash(ctx context.Context, codeHash string, now time.Time) (*entities.SsoAuthorizationCode, *utilsError.ErrorBuilder)
}

type SsoServiceI interface {
	Authorize(ctx context.Context, req models.SsoAuthorizeRequest, sessionToken string) (*models.SsoAuthorizeResponse, *utilsError.ErrorBuilder)
	Login(ctx context.Context, req models.SsoLoginRequest) (*models.SsoLoginResponse, string, time.Time, *utilsError.ErrorBuilder)
	Token(ctx context.Context, req models.SsoTokenRequest) (*models.SsoTokenResponse, *utilsError.ErrorBuilder)
	Session(ctx context.Context, sessionToken string) (*models.SsoSessionResponse, *utilsError.ErrorBuilder)
	Logout(ctx context.Context, sessionToken string) (*models.SsoLogoutResponse, *utilsError.ErrorBuilder)
}

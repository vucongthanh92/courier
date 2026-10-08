package interfaces

import (
	"context"
	"time"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type SsoSessionCommandRepoI interface {
	CreateSession(ctx context.Context, entity *entities.SsoSession) *errHandler.ErrorBuilder
	RevokeByHash(ctx context.Context, sessionHash string, revokedAt time.Time) *errHandler.ErrorBuilder
	RevokeByUser(ctx context.Context, userID uint64, revokedAt time.Time) *errHandler.ErrorBuilder
}

type SsoSessionQueryRepoI interface {
	GetActiveByHash(ctx context.Context, sessionHash string, now time.Time) (*entities.SsoSession, *errHandler.ErrorBuilder)
}

type SsoAuthorizationCodeCommandRepoI interface {
	CreateCode(ctx context.Context, entity *entities.SsoAuthorizationCode) *errHandler.ErrorBuilder
	ConsumeCode(ctx context.Context, codeHash string, consumedAt time.Time) *errHandler.ErrorBuilder
}

type SsoAuthorizationCodeQueryRepoI interface {
	GetActiveByHash(ctx context.Context, codeHash string, now time.Time) (*entities.SsoAuthorizationCode, *errHandler.ErrorBuilder)
}

type SsoServiceI interface {
	Authorize(ctx context.Context, req models.SsoAuthorizeRequest, sessionToken string) (*models.SsoAuthorizeResponse, *errHandler.ErrorBuilder)
	Login(ctx context.Context, req models.SsoLoginRequest) (*models.SsoLoginResponse, string, time.Time, *errHandler.ErrorBuilder)
	Token(ctx context.Context, req models.SsoTokenRequest) (*models.SsoTokenResponse, *errHandler.ErrorBuilder)
	Session(ctx context.Context, sessionToken string) (*models.SsoSessionResponse, *errHandler.ErrorBuilder)
	Logout(ctx context.Context, sessionToken string) (*models.SsoLogoutResponse, *errHandler.ErrorBuilder)
}

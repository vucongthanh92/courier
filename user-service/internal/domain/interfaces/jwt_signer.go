package interfaces

import (
	"context"
	"time"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type JWTSignerI interface {
	SignAccessToken(user entities.User, now time.Time, ttl time.Duration) (string, *errHandler.ErrorBuilder)
	SignAccessTokenForClient(user entities.User, now time.Time, ttl time.Duration, clientID string, sessionID string, scope string) (string, *errHandler.ErrorBuilder)
	SignIDToken(user entities.User, now time.Time, ttl time.Duration, audience string, nonce string, authTime time.Time, sessionID string) (string, *errHandler.ErrorBuilder)
}

type JWKQueryRepoI interface {
	GetActiveKey(ctx context.Context) (entities.JWKKey, *errHandler.ErrorBuilder)
	GetKeyByKid(ctx context.Context, kid string) (entities.JWKKey, *errHandler.ErrorBuilder)
}

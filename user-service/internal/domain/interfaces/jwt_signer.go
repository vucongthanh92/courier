package interfaces

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
)

type JWTSignerI interface {
	SignAccessToken(user entities.User, now time.Time, ttl time.Duration) (string, *utilsError.ErrorBuilder)
	SignAccessTokenForClient(user entities.User, now time.Time, ttl time.Duration, clientID string, sessionID string, scope string) (string, *utilsError.ErrorBuilder)
	SignIDToken(user entities.User, now time.Time, ttl time.Duration, audience string, nonce string, authTime time.Time, sessionID string) (string, *utilsError.ErrorBuilder)
}

type JWKQueryRepoI interface {
	GetActiveKey(ctx context.Context) (entities.JWKKey, *utilsError.ErrorBuilder)
	GetKeyByKid(ctx context.Context, kid string) (entities.JWKKey, *utilsError.ErrorBuilder)
}

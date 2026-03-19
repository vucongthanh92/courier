package interfaces

import (
	"time"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type JWTSignerI interface {
	SignAccessToken(user entities.User, now time.Time, ttl time.Duration) (string, *errHandler.ErrorBuilder)
}

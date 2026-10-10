package interfaces

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
)

type RefreshTokenCommandRepoI interface {
	UpsertByUserAgent(ctx context.Context, entity *entities.RefreshToken) *utilsError.ErrorBuilder
	RevokeByID(ctx context.Context, id uint64, revokedAt time.Time) *utilsError.ErrorBuilder
	RevokeByUser(ctx context.Context, userID uint64, revokedAt time.Time) *utilsError.ErrorBuilder
	Rotate(ctx context.Context, oldID uint64, newEntity *entities.RefreshToken) (*entities.RefreshToken, *utilsError.ErrorBuilder)
	DeleteExpiredAndRevoked(ctx context.Context, now time.Time) error
}

type RefreshTokenQueryRepoI interface {
	GetByTokenHash(ctx context.Context, hash string) (entities.RefreshToken, *utilsError.ErrorBuilder)
}

package interfaces

import (
	"context"
	"time"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type RefreshTokenCommandRepoI interface {
	Insert(ctx context.Context, entity *entities.RefreshToken) *errHandler.ErrorBuilder
	RevokeByID(ctx context.Context, id uint64, revokedAt time.Time) *errHandler.ErrorBuilder
	RevokeByUser(ctx context.Context, userID uint64, revokedAt time.Time) *errHandler.ErrorBuilder // optional logout-all
	Rotate(ctx context.Context, oldID uint64, newEntity *entities.RefreshToken) (*entities.RefreshToken, *errHandler.ErrorBuilder)
}

type RefreshTokenQueryRepoI interface {
	GetByTokenHash(ctx context.Context, hash string) (entities.RefreshToken, *errHandler.ErrorBuilder)
}

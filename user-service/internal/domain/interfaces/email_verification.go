package interfaces

import (
	"context"
	"time"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type EmailVerificationQueryRepoI interface {
	GetActiveByEmail(ctx context.Context, email string) (entities.EmailVerification, *errHandler.ErrorBuilder)
}

type EmailVerificationCommandRepoI interface {
	InsertEmailVerification(ctx context.Context, entity *entities.EmailVerification) *errHandler.ErrorBuilder
	UpdateToken(ctx context.Context, email string, tokenHash string, expiresAt time.Time) *errHandler.ErrorBuilder
	MarkUsed(ctx context.Context, id uint64, usedAt time.Time) *errHandler.ErrorBuilder
}

type EmailVerificationServiceI interface {
}

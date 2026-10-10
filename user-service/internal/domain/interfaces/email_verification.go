package interfaces

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
)

type EmailVerificationQueryRepoI interface {
	GetOneByEmail(ctx context.Context, email string) (entities.EmailVerification, *utilsError.ErrorBuilder)
}

type EmailVerificationCommandRepoI interface {
	InsertEmailVerification(ctx context.Context, entity *entities.EmailVerification) *utilsError.ErrorBuilder
	UpdateToken(ctx context.Context, email string, tokenHash string, expiresAt time.Time) *utilsError.ErrorBuilder
	MarkUsed(ctx context.Context, id uint64, usedAt time.Time) *utilsError.ErrorBuilder
}

type EmailVerificationServiceI interface {
}

package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type EmailVerificationQueryRepoI interface {
}

type EmailVerificationCommandRepoI interface {
	InsertEmailVerification(ctx context.Context, entity *entities.EmailVerification) *errHandler.ErrorBuilder
}

type EmailVerificationServiceI interface {
}

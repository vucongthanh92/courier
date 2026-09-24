package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/entities"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/models"
)

// IdempotencyCommandRepoI defines the interface for managing idempotency keys in the repository.
type IdempotencyCommandRepoI interface {
	Claim(context.Context, *entities.IdempotencyKey) (*entities.IdempotencyKey, bool, *errHandler.ErrorBuilder)
	Complete(context.Context, uint64, int16, []byte) *errHandler.ErrorBuilder
}

type IdempotencyServiceI interface {
	Claim(context.Context, models.IdempotencyClaimRequest) (*models.IdempotencyClaimResult, *errHandler.ErrorBuilder)
	Complete(context.Context, uint64, int16, any) *errHandler.ErrorBuilder
}

// OutboxCommandRepoI defines the interface for managing outbox events in the repository.
type OutboxCommandRepoI interface {
	Create(context.Context, *entities.OutboxEvent) *errHandler.ErrorBuilder
}

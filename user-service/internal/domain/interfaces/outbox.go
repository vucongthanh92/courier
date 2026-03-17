package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type OutboxQueryRepoI interface {
	GetOutboxByID(ctx context.Context, id uint64) (*entities.Outbox, *errHandler.ErrorBuilder)
}

type OutboxCommandRepoI interface {
	InsertOutbox(ctx context.Context, entity entities.Outbox) (entities.Outbox, *errHandler.ErrorBuilder)
	UpdateOutboxPublished(ctx context.Context, entity *entities.Outbox) *errHandler.ErrorBuilder
}

type OutboxServiceI interface {
}

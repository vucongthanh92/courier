package interfaces

import (
	"context"

	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type OutboxQueryRepoI interface {
}

type OutboxCommandRepoI interface {
	InsertOutbox(ctx context.Context, entity entities.Outbox) (
		entities.Outbox, *errHandler.ErrorBuilder)
}

type OutboxServiceI interface {
}

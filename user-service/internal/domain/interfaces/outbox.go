package interfaces

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	utilsError "github.com/vucongthanh92/go-base-utils/helper/http_error"
)

type OutboxQueryRepoI interface {
	GetOutboxByID(ctx context.Context, id uint64) (*entities.Outbox, *utilsError.ErrorBuilder)
}

type OutboxCommandRepoI interface {
	InsertOutbox(ctx context.Context, entity entities.Outbox) (entities.Outbox, *utilsError.ErrorBuilder)
	UpdateOutboxPublished(ctx context.Context, entity *entities.Outbox) *utilsError.ErrorBuilder
}

type OutboxServiceI interface {
	CreateOutbox(ctx context.Context, req models.CreateOutboxRequest) *utilsError.ErrorBuilder
}

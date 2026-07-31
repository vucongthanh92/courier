package interfaces

import (
	"context"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type IntegrationEventPublisherI interface {
	PublishOutbox(ctx context.Context, event *entities.Outbox) error
	Close() error
}

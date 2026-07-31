package interfaces

import (
	"context"

	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type WsPublisherI interface {
	PublishMessageCreated(ctx context.Context, event models.MessageCreatedEvent) error
}

type WsSubscriberI interface {
	SubscribeMessageCreated(ctx context.Context) (<-chan models.MessageCreatedEvent, <-chan error)
}

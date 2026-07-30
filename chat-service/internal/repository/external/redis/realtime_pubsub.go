package redis

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	redisClient "github.com/vucongthanh92/courier/chat-service/redis"
)

const messageCreatedChannel = "chat:events:message.created"

type realtimePubSub struct {
	client redis.UniversalClient
}

func InitRealtimePublisher(client redisClient.Client) interfaces.RealtimePublisherI {
	return &realtimePubSub{client: redis.UniversalClient(client)}
}

func InitRealtimeSubscriber(client redisClient.Client) interfaces.RealtimeSubscriberI {
	return &realtimePubSub{client: redis.UniversalClient(client)}
}

func (r *realtimePubSub) PublishMessageCreated(ctx context.Context, event models.MessageCreatedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, messageCreatedChannel, payload).Err()
}

func (r *realtimePubSub) SubscribeMessageCreated(ctx context.Context) (<-chan models.MessageCreatedEvent, <-chan error) {
	events := make(chan models.MessageCreatedEvent)
	errs := make(chan error, 1)
	pubsub := r.client.Subscribe(ctx, messageCreatedChannel)

	go func() {
		defer close(events)
		defer close(errs)
		defer pubsub.Close()

		if _, err := pubsub.Receive(ctx); err != nil {
			errs <- err
			return
		}

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var event models.MessageCreatedEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					select {
					case errs <- err:
					default:
					}
					continue
				}
				select {
				case events <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return events, errs
}

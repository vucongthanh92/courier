package redis

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	redisClient "github.com/vucongthanh92/courier/chat-service/redis"
)

type wsPubSub struct {
	client redis.UniversalClient
}

func InitWsPublisher(client redisClient.Client) interfaces.WsPublisherI {
	return &wsPubSub{client: redis.UniversalClient(client)}
}

func InitWsSubscriber(client redisClient.Client) interfaces.WsSubscriberI {
	return &wsPubSub{client: redis.UniversalClient(client)}
}

func (r *wsPubSub) PublishMessageCreated(ctx context.Context, event models.MessageCreatedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, constants.MessageCreatedChannel, payload).Err()
}

func (r *wsPubSub) SubscribeMessageCreated(ctx context.Context) (<-chan models.MessageCreatedEvent, <-chan error) {
	events := make(chan models.MessageCreatedEvent)
	errs := make(chan error, 1)
	pubsub := r.client.Subscribe(ctx, constants.MessageCreatedChannel)

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

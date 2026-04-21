package dispatcher

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
	"gorm.io/gorm"
)

// Publisher is your broker adapter (Kafka/NATS/etc.)
type Publisher interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

// Dispatcher pulls un-published outbox rows and publishes them
// Run in a background goroutine / worker
func RunDispatcher(ctx context.Context, db *gorm.DB, pub Publisher, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_ = dispatchOnce(ctx, db, pub)
		}
	}
}

func dispatchOnce(ctx context.Context, db *gorm.DB, pub Publisher) error {
	var msgs []entities.Outbox

	if err := db.WithContext(ctx).
		Where("published_at IS NULL").
		Order("created_at ASC").
		Limit(100).
		Find(&msgs).Error; err != nil {
		return err
	}

	if len(msgs) == 0 {
		return nil
	}

	for _, msg := range msgs {
		topic := msg.AggregateType + "." + msg.EventType
		key := msg.AggregateID
		if err := pub.Publish(ctx, topic, key, msg.Payload); err != nil {
			continue
		}
	}

	return nil
}

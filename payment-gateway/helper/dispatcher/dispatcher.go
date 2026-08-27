package dispatcher

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

func RunDispatcher(ctx context.Context, _ *gorm.DB, _ Publisher, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

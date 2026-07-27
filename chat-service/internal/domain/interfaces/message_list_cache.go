package interfaces

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

type MessageListCacheI interface {
	GetLatest(ctx context.Context, conversationID uint64, limit int) (*models.CachedMessageListPage, error)
	SetLatest(ctx context.Context, conversationID uint64, limit int, page models.CachedMessageListPage, ttl time.Duration) error
	InvalidateLatest(ctx context.Context, conversationID uint64) error
}

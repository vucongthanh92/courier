package interfaces

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

// UserProfileCacheI defines the interface for caching user profile summaries.
type UserProfileCacheI interface {
	GetMany(ctx context.Context, userIDs []uint64) (map[uint64]models.UserProfileSummaryResponse, error)
	SetMany(ctx context.Context, profiles []models.UserProfileSummaryResponse, ttl time.Duration) error
}

// MessageRateLimiterI defines the interface for rate limiting message sending in conversations.
type MessageRateLimitResult struct {
	Allowed    bool
	Limit      int
	Remaining  int
	RetryAfter int
	ResetAt    int64
}

type MessageRateLimiterI interface {
	Allow(ctx context.Context, userID, conversationID uint64) (MessageRateLimitResult, error)
}

// MessageListCacheI defines the interface for caching the latest messages in a conversation.
type MessageListCacheI interface {
	GetLatest(ctx context.Context, conversationID uint64, limit int) (*models.CachedMessageListPage, error)
	SetLatest(ctx context.Context, conversationID uint64, limit int, page models.CachedMessageListPage, ttl time.Duration) error
	InvalidateLatest(ctx context.Context, conversationID uint64) error
}

// TokenDenylistI defines the interface for managing a denylist of JWT tokens.
type TokenDenylistI interface {
	Block(ctx context.Context, jti string, ttl time.Duration) error
	IsBlocked(ctx context.Context, jti string) (bool, error)
}

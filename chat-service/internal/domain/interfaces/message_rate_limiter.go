package interfaces

import "context"

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

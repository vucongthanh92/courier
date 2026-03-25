package interfaces

import (
	"context"
	"time"
)

type TokenDenylistI interface {
	Block(ctx context.Context, jti string, ttl time.Duration) error
	IsBlocked(ctx context.Context, jti string) (bool, error)
}

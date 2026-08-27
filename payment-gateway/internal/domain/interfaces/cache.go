package interfaces

import (
	"context"
	"time"
)

type TokenDenylistI interface {
	Block(context.Context, string, time.Duration) error
	IsBlocked(context.Context, string) (bool, error)
}

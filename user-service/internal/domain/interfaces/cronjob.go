package interfaces

import "context"

type CronJobServiceI interface {
	CleanupRefreshTokens(ctx context.Context) error
}

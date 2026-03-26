package cronjob

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

type CronJobImpl struct {
	refreshTokenCmd interfaces.RefreshTokenCommandRepoI
}

func NewCronJobService(
	refreshTokenCmd interfaces.RefreshTokenCommandRepoI,
) interfaces.CronJobServiceI {
	return &CronJobImpl{
		refreshTokenCmd: refreshTokenCmd,
	}
}

// CleanupRefreshTokens deletes expired and revoked refresh tokens from the database.
func (s *CronJobImpl) CleanupRefreshTokens(ctx context.Context) error {
	// Get the current time to use as a cutoff for deleting expired tokens
	cutoff := time.Now()
	return s.refreshTokenCmd.DeleteExpiredAndRevoked(ctx, cutoff)
}

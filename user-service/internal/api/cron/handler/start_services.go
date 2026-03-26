package handler

import (
	"context"

	"github.com/jasonlvhit/gocron"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

// StartServices starts the cron jobs for the application.
func StartServices(cronService interfaces.CronJobServiceI) {
	gocron.Every(1).Minutes().Do(func() {
		_ = cronService.CleanupRefreshTokens(context.Background())
	})
	go gocron.Start()
}

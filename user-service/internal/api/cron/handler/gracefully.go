package handler

import (
	"github.com/vucongthanh92/courier/user-service/config"
	"github.com/vucongthanh92/courier/user-service/internal/application/cronjob"
)

func Gracefully(cfg *config.AppConfig, cronService cronjob.CronJobService) (err error) {
	return nil
}

package handler

import (
	"github.com/vucongthanh92/go-base-project/config"
	"github.com/vucongthanh92/go-base-project/internal/application/cronjob"
)

func Gracefully(cfg *config.AppConfig, cronService cronjob.CronJobService) (err error) {
	return nil
}

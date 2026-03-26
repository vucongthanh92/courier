package handler

import (
	"github.com/vucongthanh92/courier/user-service/config"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
)

func Gracefully(cfg *config.AppConfig, cronService interfaces.CronJobServiceI) (err error) {
	return nil
}

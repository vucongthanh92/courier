package cron

import (
	"github.com/vucongthanh92/go-base-project/config"
	"github.com/vucongthanh92/go-base-project/internal/api/cron/handler"
	"github.com/vucongthanh92/go-base-project/internal/application/cronjob"

	"github.com/vucongthanh92/go-base-utils/logger"
)

type Server struct {
	cfg         *config.AppConfig
	cronService cronjob.CronJobService
}

func NewServer(
	cfg *config.AppConfig,
	cronService cronjob.CronJobService,
) *Server {
	return &Server{
		cfg:         cfg,
		cronService: cronService,
	}
}

func (s *Server) Run() {
	if s.cfg.CronJob.Disable {
		logger.Info(`CronJob Not Enable`)
		return
	}

	// Start the scheduler
	handler.Crawl(
		s.cfg,
		s.cronService,
	)
}

func init() {
	logger.Info("Init service cron ...")
}

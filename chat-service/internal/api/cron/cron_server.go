package cron

import "github.com/vucongthanh92/courier/chat-service/config"

type Server struct {
	cfg *config.AppConfig
}

func NewServer(cfg *config.AppConfig) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Run() {
	if s.cfg.CronJob != nil && s.cfg.CronJob.Disable {
		return
	}
}

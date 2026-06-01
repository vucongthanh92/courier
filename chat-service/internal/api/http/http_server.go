package http

import (
	"github.com/vucongthanh92/courier/chat-service/config"
	v1 "github.com/vucongthanh92/courier/chat-service/internal/api/http/v1"
	httpserver "github.com/vucongthanh92/go-base-utils/http/server"
)

type Server struct {
	cfg                 *config.AppConfig
	conversationHandler *v1.ConversationHandler
}

func NewServer(cfg *config.AppConfig, conversationHandler *v1.ConversationHandler) *Server {
	return &Server{
		cfg:                 cfg,
		conversationHandler: conversationHandler,
	}
}

func (s *Server) Run() {
	serverConfig := &httpserver.HttpServerConfig{
		Port:            s.cfg.Http.Port,
		Development:     s.cfg.Http.Development,
		ShutdownTimeout: s.cfg.Http.ShutdownTimeout,
		Resources:       s.cfg.Http.Resources,
		AllowOrigins:    s.cfg.Http.AllowOrigins,
	}

	httpServer, router := httpserver.NewServer(*serverConfig)
	v1.MapRoutes(router, s.conversationHandler)

	httpServer.Run()
}

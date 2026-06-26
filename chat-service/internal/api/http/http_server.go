package http

import (
	"context"

	"github.com/vucongthanh92/courier/chat-service/config"
	middleware "github.com/vucongthanh92/courier/chat-service/internal/api/http/middleware"
	v1 "github.com/vucongthanh92/courier/chat-service/internal/api/http/v1"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	httpserver "github.com/vucongthanh92/go-base-utils/http/server"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

type Server struct {
	cfg                 *config.AppConfig
	conversationHandler *v1.ConversationHandler
	jwkRepo             interfaces.JWKQueryRepoI
	tokenDeny           interfaces.TokenDenylistI
}

func NewServer(
	cfg *config.AppConfig,
	conversationHandler *v1.ConversationHandler,
	jwkRepo interfaces.JWKQueryRepoI,
	tokenDeny interfaces.TokenDenylistI) *Server {
	return &Server{
		cfg:                 cfg,
		conversationHandler: conversationHandler,
		jwkRepo:             jwkRepo,
		tokenDeny:           tokenDeny,
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

	// Load public keys for JWT middleware. Currently we only support 1 active key,
	// but we design the return type as map to support multiple keys in the future without changing middleware code.
	// If we have multiple keys in the future, we can enhance LoadPubKeys to return map of kid->key and let middleware pick the right key based on "kid" in JWT header
	pubKeys, err := middleware.LoadPubKeys(context.Background(), s.jwkRepo)
	if err != nil {
		logger.Fatal("load public key failed", zap.Error(err.LogError))
	}
	authMW := middleware.JWTMiddleware(s.tokenDeny, pubKeys)

	// In the future, if we have v2, v3..., we will add at here
	v1.MapRoutes(
		router,
		s.conversationHandler,
		authMW,
	)

	httpServer.Run()
}

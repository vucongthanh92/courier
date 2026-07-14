package http

import (
	"context"

	"github.com/vucongthanh92/courier/chat-service/config"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	middleware "github.com/vucongthanh92/courier/chat-service/internal/api/http/middleware"
	v1 "github.com/vucongthanh92/courier/chat-service/internal/api/http/v1"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	jwkclient "github.com/vucongthanh92/courier/chat-service/internal/repository/external/jwkclient"
	cacheRepo "github.com/vucongthanh92/courier/chat-service/internal/repository/external/redis"
	httpserver "github.com/vucongthanh92/go-base-utils/http/server"
)

type Server struct {
	cfg                 *config.AppConfig
	conversationHandler *v1.ConversationHandler
	jwkCache            cacheRepo.JWKCacheRepo
	jwkClient           jwkclient.Client
	tokenDeny           interfaces.TokenDenylistI
}

func NewServer(
	cfg *config.AppConfig,
	conversationHandler *v1.ConversationHandler,
	jwkCache cacheRepo.JWKCacheRepo,
	jwkClient jwkclient.Client,
	tokenDeny interfaces.TokenDenylistI) *Server {
	return &Server{
		cfg:                 cfg,
		conversationHandler: conversationHandler,
		jwkCache:            jwkCache,
		jwkClient:           jwkClient,
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

	// Set up JWT middleware with denylist and public key resolver
	authMW := middleware.JWTMiddleware(s.tokenDeny, func(ctx context.Context, kid string) (interface{}, *errHandler.ErrorBuilder) {
		return middleware.ResolvePublicKey(ctx, s.jwkCache, s.jwkClient, kid)
	})

	// In the future, if we have v2, v3..., we will add at here
	v1.MapRoutes(
		router,
		s.conversationHandler,
		authMW,
	)

	httpServer.Run()
}

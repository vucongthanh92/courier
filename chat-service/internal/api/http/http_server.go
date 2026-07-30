package http

import (
	"context"

	"github.com/vucongthanh92/courier/chat-service/config"
	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
	middleware "github.com/vucongthanh92/courier/chat-service/internal/api/http/middleware"
	v1 "github.com/vucongthanh92/courier/chat-service/internal/api/http/v1"
	"github.com/vucongthanh92/courier/chat-service/internal/api/ws"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	cacheRepo "github.com/vucongthanh92/courier/chat-service/internal/repository/external/redis"
	user_grpc "github.com/vucongthanh92/courier/chat-service/internal/repository/external/user_grpc"
	httpmiddlewares "github.com/vucongthanh92/go-base-utils/http/middlewares"
	httpserver "github.com/vucongthanh92/go-base-utils/http/server"
)

type Server struct {
	cfg                 *config.AppConfig
	conversationHandler *v1.ConversationHandler
	messageHandler      *v1.MessageHandler
	realtimeHub         *ws.Hub
	messageRateLimiter  interfaces.MessageRateLimiterI
	jwkCache            cacheRepo.JWKCacheRepo
	userGrpc            user_grpc.UserGrpcClient
	tokenDeny           interfaces.TokenDenylistI
}

func NewServer(
	cfg *config.AppConfig,
	conversationHandler *v1.ConversationHandler,
	messageHandler *v1.MessageHandler,
	realtimeHub *ws.Hub,
	messageRateLimiter interfaces.MessageRateLimiterI,
	jwkCache cacheRepo.JWKCacheRepo,
	userGrpc user_grpc.UserGrpcClient,
	tokenDeny interfaces.TokenDenylistI) *Server {
	return &Server{
		cfg:                 cfg,
		conversationHandler: conversationHandler,
		messageHandler:      messageHandler,
		realtimeHub:         realtimeHub,
		messageRateLimiter:  messageRateLimiter,
		jwkCache:            jwkCache,
		userGrpc:            userGrpc,
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
	router.Use(httpmiddlewares.Cors(s.cfg.Http.AllowOrigins...))

	keyResolver := func(ctx context.Context, kid string) (interface{}, *errHandler.ErrorBuilder) {
		return middleware.ResolvePublicKey(ctx, s.jwkCache, s.userGrpc, kid)
	}
	authMW := middleware.JWTMiddleware(s.tokenDeny, keyResolver)
	if s.realtimeHub != nil {
		go s.realtimeHub.Run(context.Background())
	}
	realtimeHandler := v1.InitRealtimeHandler(s.realtimeHub, s.cfg.Http.AllowOrigins, s.tokenDeny, keyResolver)

	// In the future, if we have v2, v3..., we will add at here
	v1.MapRoutes(
		router,
		s.conversationHandler,
		s.messageHandler,
		realtimeHandler,
		authMW,
		middleware.MessageRateLimitMiddleware(s.messageRateLimiter),
	)

	httpServer.Run()
}

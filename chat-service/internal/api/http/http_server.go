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
	memberHandler       *v1.MemberHandler
	messageHandler      *v1.MessageHandler
	wsHub               *ws.Hub
	messageRateLimiter  interfaces.MessageRateLimiterI
	jwkCache            cacheRepo.JWKCacheRepo
	userGrpc            user_grpc.UserGrpcClient
	tokenDeny           interfaces.TokenDenylistI
}

func NewServer(
	cfg *config.AppConfig,
	conversationHandler *v1.ConversationHandler,
	memberHandler *v1.MemberHandler,
	messageHandler *v1.MessageHandler,
	wsHub *ws.Hub,
	messageRateLimiter interfaces.MessageRateLimiterI,
	jwkCache cacheRepo.JWKCacheRepo,
	userGrpc user_grpc.UserGrpcClient,
	tokenDeny interfaces.TokenDenylistI) *Server {
	return &Server{
		cfg:                 cfg,
		conversationHandler: conversationHandler,
		memberHandler:       memberHandler,
		messageHandler:      messageHandler,
		wsHub:               wsHub,
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
	if s.wsHub != nil {
		go s.wsHub.Run(context.Background())
	}
	wsHandler := v1.InitWsHandler(s.wsHub, s.cfg.Http.AllowOrigins, s.tokenDeny, keyResolver)

	// In the future, if we have v2, v3..., we will add at here
	v1.MapRoutes(
		router,
		s.conversationHandler,
		s.memberHandler,
		s.messageHandler,
		wsHandler,
		authMW,
		middleware.MessageRateLimitMiddleware(s.messageRateLimiter),
	)

	httpServer.Run()
}

package http

import (
	"context"

	"github.com/vucongthanh92/courier/payment-gateway/config"
	errHandler "github.com/vucongthanh92/courier/payment-gateway/helper/error_handler"
	"github.com/vucongthanh92/courier/payment-gateway/internal/api/http/middleware"
	v1 "github.com/vucongthanh92/courier/payment-gateway/internal/api/http/v1"
	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	cacheRepo "github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/redis"
	usergrpc "github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/user_grpc"
	httpmiddlewares "github.com/vucongthanh92/go-base-utils/http/middlewares"
	httpserver "github.com/vucongthanh92/go-base-utils/http/server"
)

// Server owns HTTP bootstrap only. Versioned routes and request handling live in v1.
type Server struct {
	cfg                 *config.AppConfig
	topUpHandler        *v1.TopUpHandler
	sePayWebhookHandler *v1.SePayWebhookHandler
	jwkCache            cacheRepo.JWKCacheRepo
	userGrpc            usergrpc.UserGrpcClient
	tokenDeny           interfaces.TokenDenylistI
}

func NewServer(
	cfg *config.AppConfig,
	topUpHandler *v1.TopUpHandler,
	sePayWebhookHandler *v1.SePayWebhookHandler,
	jwkCache cacheRepo.JWKCacheRepo,
	userGrpc usergrpc.UserGrpcClient,
	tokenDeny interfaces.TokenDenylistI,
) *Server {
	return &Server{
		cfg:                 cfg,
		topUpHandler:        topUpHandler,
		sePayWebhookHandler: sePayWebhookHandler,
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

	// In the future, if we have v2, v3..., we will add at here
	v1.MapRoutes(
		router,
		s.topUpHandler,
		s.sePayWebhookHandler,
		authMW,
	)

	httpServer.Run()
}

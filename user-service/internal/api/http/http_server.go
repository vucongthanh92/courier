package http

import (
	"context"
	"os"

	"github.com/swaggo/swag"
	"github.com/vucongthanh92/courier/user-service/config"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	middleware "github.com/vucongthanh92/courier/user-service/internal/api/http/middleware"
	v1 "github.com/vucongthanh92/courier/user-service/internal/api/http/v1"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	cacheRepo "github.com/vucongthanh92/courier/user-service/internal/repository/external/redis"
	httpmiddlewares "github.com/vucongthanh92/go-base-utils/http/middlewares"
	httpserver "github.com/vucongthanh92/go-base-utils/http/server"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

type Server struct {
	cfg               *config.AppConfig
	authHandler       *v1.AuthHandler
	identityHandler   *v1.IdentityHandler
	credentialHandler *v1.CredentialHandler
	jwkRepo           interfaces.JWKQueryRepoI
	tokenDeny         interfaces.TokenDenylistI
	jwkCache          cacheRepo.JWKCacheRepo
}

func NewServer(
	cfg *config.AppConfig,
	authHandler *v1.AuthHandler,
	identityHandler *v1.IdentityHandler,
	credentialHandler *v1.CredentialHandler,
	jwkRepo interfaces.JWKQueryRepoI,
	tokenDeny interfaces.TokenDenylistI,
	jwkCache cacheRepo.JWKCacheRepo,
) *Server {
	return &Server{
		cfg:               cfg,
		authHandler:       authHandler,
		identityHandler:   identityHandler,
		credentialHandler: credentialHandler,
		jwkRepo:           jwkRepo,
		tokenDeny:         tokenDeny,
		jwkCache:          jwkCache,
	}
}

func (s *Server) Run() {
	config := &httpserver.HttpServerConfig{
		Port:            s.cfg.Http.Port,
		Development:     s.cfg.Http.Development,
		ShutdownTimeout: s.cfg.Http.ShutdownTimeout,
		Resources:       s.cfg.Http.Resources,
		AllowOrigins:    s.cfg.Http.AllowOrigins,
	}
	httpServer, router := httpserver.NewServer(*config)
	router.Use(httpmiddlewares.Cors(s.cfg.Http.AllowOrigins...))

	// // Add recover panic middleware
	// router.Use(middlewares.RecoverPanicMiddleware(middlewares.RecoverPanicMiddlewareConfig{
	// 	SlackConfig: slack.SlackConfig{
	// 		Channel:         s.cfg.SlackService.Channel,
	// 		Username:        s.cfg.SlackService.Username,
	// 		UrlSlackWebHook: s.cfg.SlackService.UrlSlackWebhook,
	// 	},
	// }))

	// Load public keys for JWT middleware. Currently we only support 1 active key,
	// but we design the return type as map to support multiple keys in the future without changing middleware code.
	// If we have multiple keys in the future, we can enhance LoadPubKeys to return map of kid->key and let middleware pick the right key based on "kid" in JWT header
	if _, err := middleware.LoadPubKeys(context.Background(), s.jwkRepo, s.jwkCache); err != nil {
		logger.Fatal("load public key failed", zap.Error(err.LogError))
	}
	authMW := middleware.JWTMiddleware(s.tokenDeny, func(ctx context.Context, kid string) (any, *errHandler.ErrorBuilder) {
		return middleware.ResolvePublicKey(ctx, s.jwkRepo, s.jwkCache, kid)
	})

	// In the future, if we have v2, v3..., we will add at here
	v1.MapRoutes(
		router,
		s.authHandler,
		s.credentialHandler,
		s.identityHandler,
		authMW,
	)
	httpServer.Run()
}

// For Swagger docs, we read the generated swagger.json to avoid issues
// with swaggo annotations and maintain separation of concerns between code and docs.
func init() {
	dat, err := os.ReadFile("./docs/swagger.json")
	if err != nil {
		println("error when reading specs, please regenerate swagger")
	}
	spec := &swag.Spec{
		Version:          "1.0",
		BasePath:         "/api/v1/",
		Schemes:          []string{},
		Title:            "User Service API",
		Description:      "Service for user related api",
		InfoInstanceName: "swagger",
		SwaggerTemplate:  string(dat),
	}
	swag.Register(spec.InstanceName(), spec)
}

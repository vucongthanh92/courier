package http

import (
	"os"

	"github.com/swaggo/swag"
	"github.com/vucongthanh92/courier/user-service/config"
	v1 "github.com/vucongthanh92/courier/user-service/internal/api/http/v1"
	httpserver "github.com/vucongthanh92/go-base-utils/http/server"
)

type Server struct {
	cfg             *config.AppConfig
	userHandler     *v1.UserHandler
	identityHandler *v1.IdentityHandler
}

func NewServer(
	cfg *config.AppConfig,
	userHandler *v1.UserHandler,
	identityHandler *v1.IdentityHandler,
) *Server {
	return &Server{
		cfg:             cfg,
		userHandler:     userHandler,
		identityHandler: identityHandler,
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

	// // Add recover panic middleware
	// router.Use(middlewares.RecoverPanicMiddleware(middlewares.RecoverPanicMiddlewareConfig{
	// 	SlackConfig: slack.SlackConfig{
	// 		Channel:         s.cfg.SlackService.Channel,
	// 		Username:        s.cfg.SlackService.Username,
	// 		UrlSlackWebHook: s.cfg.SlackService.UrlSlackWebhook,
	// 	}}))

	// In the future, if we have v2, v3..., we will add at here
	v1.MapRoutes(
		router,
		s.userHandler,
		s.identityHandler,
	)
	httpServer.Run()
}

func init() {
	dat, err := os.ReadFile("./docs/swagger.json")
	if err != nil {
		println("error when reading specs, please regenerate swagger")
	}
	spec := &swag.Spec{
		Version:          "1.0",
		BasePath:         "/api/v1/",
		Schemes:          []string{},
		Title:            "Order Service API",
		Description:      "Service for order related api",
		InfoInstanceName: "swagger",
		SwaggerTemplate:  string(dat),
	}
	swag.Register(spec.InstanceName(), spec)
}

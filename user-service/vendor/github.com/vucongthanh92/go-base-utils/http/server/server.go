package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	midd "github.com/vucongthanh92/go-base-utils/http/middlewares"

	"github.com/vucongthanh92/go-base-utils/logger"

	"github.com/vucongthanh92/go-base-utils/metrics"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	log                logger.Logger
	cfg                HttpServerConfig
	Router             *gin.Engine
	httpServerInstance *http.Server
}

type HttpServer interface {
	Run()
	Stop()
}

func NewServer(cfg HttpServerConfig, options ...HttpServerOption) (HttpServer, *gin.Engine) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	httpServerInstance := &http.Server{
		Addr:    cfg.Port,
		Handler: router,
	}
	instance := &Server{cfg: cfg, Router: router, httpServerInstance: httpServerInstance}
	for _, option := range options {
		option(instance)
	}

	if instance.log == nil {
		instance.log = logger.GetDefaultLogger()
	}
	router.Use(midd.SetLanguage(cfg.Resources))
	router.Use(midd.Gzip())
	router.Use(midd.RequestId())
	router.Use(midd.Tracing(cfg.Name))
	router.Use(midd.TraceIDMiddleware())
	router.Use(midd.Recovery(instance.log))
	metrics.Use(router)
	//auth token is error when logging don't know root cause yet
	router.Use(midd.Logging())
	if cfg.RateLimiting != nil {
		router.Use(midd.RateLimiting(instance.log, router, cfg.RateLimiting.RateFormat))
	}
	return instance, router
}

func (s *Server) Run() {
	go func() {
		s.log.Info("Http server is listening at: ", zap.String("PORT", s.cfg.Port))
		if err := s.httpServerInstance.ListenAndServe(); err != nil {
			s.log.Error("failed to listen", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	s.log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := s.httpServerInstance.Shutdown(ctx); err != nil {
		s.log.Error("Server Shutdown:", zap.Error(err))
	}

	close(quit)
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.ShutdownTimeout)*time.Second)
	defer cancel()
	if err := s.httpServerInstance.Shutdown(ctx); err != nil {
		s.log.Fatal("Server Shutdown:", zap.Error(err))
	}
}

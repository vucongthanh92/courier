package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vucongthanh92/courier/agent-gateway/config"
	"github.com/vucongthanh92/courier/agent-gateway/helper/constants"
	"github.com/vucongthanh92/courier/agent-gateway/internal/gateway"
	"github.com/vucongthanh92/courier/agent-gateway/internal/repository/qdrant"
)

func main() {
	configPath := flag.String("config", "./config/local/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	qdrantClient := qdrant.NewClient(cfg.Qdrant)
	agentGateway := gateway.NewService(cfg, qdrantClient)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := agentGateway.Bootstrap(ctx); err != nil {
		log.Fatalf("bootstrap agent-gateway: %v", err)
	}

	server := &http.Server{
		Addr:              cfg.HTTP.Port,
		Handler:           gateway.NewHTTPHandler(agentGateway),
		ReadHeaderTimeout: constants.DefaultReadHeaderTimeout,
	}

	go func() {
		log.Printf("agent-gateway listening on %s", cfg.HTTP.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("run agent-gateway: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown agent-gateway: %v", err)
	}
}

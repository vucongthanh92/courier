package handler

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jasonlvhit/gocron"
	"github.com/vucongthanh92/courier/user-service/config"
	"github.com/vucongthanh92/courier/user-service/helper/utils"
	"github.com/vucongthanh92/courier/user-service/internal/application/cronjob"
	"github.com/vucongthanh92/go-base-utils/logger"
)

func Crawl(
	cfg *config.AppConfig,
	cronService cronjob.CronJobService,
) {
	ctx, cancel := context.WithCancel(context.Background())
	shutdownCh := make(chan struct{})

	// Listen for interrupt signals to initiate graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		logger.Info("Received interrupt signal. Shutting down gracefully...")
		close(shutdownCh) // Signal the shutdown
		cancel()
	}()

	utils.SafeGo(func() {
		StartServices(cronService)
	})

	// Wait for the context to be canceled (either by the scheduler or interrupt signal)
	<-ctx.Done()
	// Ensure that the scheduler is stopped before updating UserJob
	gocron.Clear()

	if err := Gracefully(cfg, cronService); err != nil {
		logger.Error(`error endGracefully`)
		return
	}

	logger.Info("Shutting down Crawl Data gracefully...")
}

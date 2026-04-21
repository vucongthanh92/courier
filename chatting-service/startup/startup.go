package startup

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gammazero/workerpool"
	"github.com/vucongthanh92/courier/chat-service/config"
	"github.com/vucongthanh92/courier/chat-service/database"
	"github.com/vucongthanh92/courier/chat-service/internal"
	"github.com/vucongthanh92/courier/chat-service/internal/api"
	"github.com/vucongthanh92/courier/chat-service/redis"
	"github.com/vucongthanh92/go-base-utils/command"
	"github.com/vucongthanh92/go-base-utils/logger"
)

var cfg *config.AppConfig

func Execute() {
	cfg = &config.AppConfig{}

	command.UseCommands(
		command.WithStartCommand(start, cfg),
	)
}

func start() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	container, _, _ := registerDependencies(ctx)
	runServer(container)
}

func registerDependencies(_ context.Context) (*api.ApiContainer, database.GormReadDb, database.GormWriteDb) {
	redisClient := redis.Open(cfg.Redis)
	readDb, writeDb := database.GetConnectByGorm(cfg.Database)

	return internal.InitializeContainer(
		cfg,
		&readDb,
		&writeDb,
		redisClient,
	), readDb, writeDb
}

func runServer(container *api.ApiContainer) {
	wp := workerpool.New(3)

	wp.Submit(container.GrpcServer.Run)
	wp.Submit(container.HttpServer.Run)
	wp.Submit(container.CronServer.Run)

	wp.StopWait()
	logger.Info("chat-service stopped")
}

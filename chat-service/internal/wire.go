//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/google/wire"
	"github.com/vucongthanh92/courier/chat-service/config"
	"github.com/vucongthanh92/courier/chat-service/database"
	"github.com/vucongthanh92/courier/chat-service/helper/transaction"
	"github.com/vucongthanh92/courier/chat-service/internal/api"
	"github.com/vucongthanh92/courier/chat-service/internal/api/cron"
	grpc "github.com/vucongthanh92/courier/chat-service/internal/api/grpc"
	"github.com/vucongthanh92/courier/chat-service/internal/api/http"
	"github.com/vucongthanh92/courier/chat-service/internal/api/ws"
	"github.com/vucongthanh92/courier/chat-service/internal/worker"
	"github.com/vucongthanh92/courier/chat-service/redis"

	kafkaRepo "github.com/vucongthanh92/courier/chat-service/internal/repository/external/kafka"
	cacheRepo "github.com/vucongthanh92/courier/chat-service/internal/repository/external/redis"
	user_grpc "github.com/vucongthanh92/courier/chat-service/internal/repository/external/user_grpc"

	conversationRepo "github.com/vucongthanh92/courier/chat-service/internal/repository/persistent/conversation"
	memberRepo "github.com/vucongthanh92/courier/chat-service/internal/repository/persistent/member"
	messageRepo "github.com/vucongthanh92/courier/chat-service/internal/repository/persistent/message"

	conversationUc "github.com/vucongthanh92/courier/chat-service/internal/usecase/conversation"
	memberUc "github.com/vucongthanh92/courier/chat-service/internal/usecase/member"
	messageUc "github.com/vucongthanh92/courier/chat-service/internal/usecase/message"

	v1 "github.com/vucongthanh92/courier/chat-service/internal/api/http/v1"
	"github.com/vucongthanh92/go-base-utils/logger"
)

var container = wire.NewSet(
	api.NewApiContainer,
)

var apiSet = wire.NewSet(
	cron.NewServer,
	grpc.NewServer,
	http.NewServer,
)

var handlerSet = wire.NewSet(
	v1.InitConversationHandler,
	v1.InitMemberHandler,
	v1.InitMessageHandler,
)

var serviceSet = wire.NewSet(
	conversationUc.InitConversationUsecase,
	memberUc.InitMemberUsecase,
	messageUc.InitMessageUsecase,
	conversationUc.InitUserEventHandler,
)

var repoSet = wire.NewSet(
	conversationRepo.InitConversationCommandRepo,
	conversationRepo.InitConversationQueryRepo,
	memberRepo.InitMemberCommandRepo,
	memberRepo.InitMemberQueryRepo,
	messageRepo.InitMessageCmdRepo,
	messageRepo.InitMessageQueryRepo,
	cacheRepo.InitJWKCacheRepo,
	cacheRepo.InitRedisDenylist,
	cacheRepo.InitMessageRateLimiter,
	cacheRepo.InitMessageListCache,
	cacheRepo.InitUserProfileCache,
	cacheRepo.InitWsPublisher,
	cacheRepo.InitWsSubscriber,
	kafkaRepo.InitAssistantEventPublisher,
)

var providerSet = wire.NewSet(
	user_grpc.NewGrpcClient,
	transaction.InitManagerTxn,
	ws.NewHub,
	worker.InitUserEventConsumer,
	worker.InitAssistantResponseConsumer,
	provideLogger,
)

func InitializeContainer(
	appCfg *config.AppConfig,
	readDb *database.GormReadDb,
	writeDb *database.GormWriteDb,
	redisClient redis.Client,
) *api.ApiContainer {
	wire.Build(repoSet, serviceSet, handlerSet, apiSet, providerSet, container)
	return &api.ApiContainer{}
}

// func provideUserGrpcClient(cfg *config.AppConfig) user_grpc.Client {
// 	client, err := user_grpc.NewClient(cfg.Client.UserService)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return client
// }

func provideLogger(cfg *config.AppConfig) logger.Logger {
	return logger.NewZapLogger(cfg.Logger.LogLevel)
}

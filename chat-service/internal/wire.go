//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/google/wire"
	"github.com/vucongthanh92/courier/chat-service/config"
	"github.com/vucongthanh92/courier/chat-service/database"
	"github.com/vucongthanh92/courier/chat-service/internal/api"
	"github.com/vucongthanh92/courier/chat-service/internal/api/cron"
	grpc "github.com/vucongthanh92/courier/chat-service/internal/api/grpc"
	"github.com/vucongthanh92/courier/chat-service/internal/api/http"
	"github.com/vucongthanh92/courier/chat-service/redis"

	jwkclient "github.com/vucongthanh92/courier/chat-service/internal/repository/external/jwkclient"
	cacheRepo "github.com/vucongthanh92/courier/chat-service/internal/repository/external/redis"
	conversationRepo "github.com/vucongthanh92/courier/chat-service/internal/repository/persistent/conversation"

	conversationUc "github.com/vucongthanh92/courier/chat-service/internal/usecase/conversation"

	v1 "github.com/vucongthanh92/courier/chat-service/internal/api/http/v1"
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
)

var serviceSet = wire.NewSet(
	conversationUc.InitConversationUsecase,
)

var repoSet = wire.NewSet(
	conversationRepo.InitConversationCommandRepo,
	conversationRepo.InitConversationQueryRepo,
	cacheRepo.InitJWKCacheRepo,
	cacheRepo.InitRedisDenylist,
)

var providerSet = wire.NewSet(
	provideJWKClient,
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

func provideJWKClient(cfg *config.AppConfig) jwkclient.Client {
	client, err := jwkclient.New(cfg.Client.UserService)
	if err != nil {
		panic(err)
	}
	return client
}

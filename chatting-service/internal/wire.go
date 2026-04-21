//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/google/wire"
	"github.com/vucongthanh92/courier/chat-service/config"
	"github.com/vucongthanh92/courier/chat-service/database"
	"github.com/vucongthanh92/courier/chat-service/internal/api"
	chatcron "github.com/vucongthanh92/courier/chat-service/internal/api/cron"
	chatgrpc "github.com/vucongthanh92/courier/chat-service/internal/api/grpc"
	chathttp "github.com/vucongthanh92/courier/chat-service/internal/api/http"
	"github.com/vucongthanh92/courier/chat-service/redis"
)

var containerSet = wire.NewSet(
	api.NewApiContainer,
	chathttp.NewServer,
	chatgrpc.NewServer,
	chatcron.NewServer,
)

func InitializeContainer(
	appCfg *config.AppConfig,
	readDb *database.GormReadDb,
	writeDb *database.GormWriteDb,
	redisClient redis.Client,
) *api.ApiContainer {
	_ = appCfg
	_ = readDb
	_ = writeDb
	_ = redisClient

	wire.Build(containerSet)
	return &api.ApiContainer{}
}

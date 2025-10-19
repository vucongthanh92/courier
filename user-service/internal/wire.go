//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/google/wire"
	"github.com/vucongthanh92/courier/user-service/config"
	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/courier/user-service/internal/api"
	"github.com/vucongthanh92/courier/user-service/internal/api/cron"
	"github.com/vucongthanh92/courier/user-service/internal/api/http"
	"github.com/vucongthanh92/courier/user-service/internal/application/cronjob"
	"github.com/vucongthanh92/courier/user-service/redis"

	identityUc "github.com/vucongthanh92/courier/user-service/internal/application/identity"
	userUc "github.com/vucongthanh92/courier/user-service/internal/application/user"

	identityRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/identity"
	userRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/user"

	grpcserver "github.com/vucongthanh92/courier/user-service/internal/api/grpc"
	v1 "github.com/vucongthanh92/courier/user-service/internal/api/http/v1"
)

var container = wire.NewSet(
	api.NewApiContainer,
)

var apiSet = wire.NewSet(
	cron.NewServer,
	grpcserver.NewServer,
	http.NewServer,
)

var handlerSet = wire.NewSet(
	v1.InitIdentityHandler,
	v1.InitUserHandler,
)

var serviceSet = wire.NewSet(
	cronjob.NewCronJobService,
	userUc.InitUserService,
	identityUc.InitIdentityService,
)

var repoSet = wire.NewSet(
	userRepo.InitUserCmdRepository,
	userRepo.InitUserQueryRepository,
	identityRepo.InitIdentityCmdRepository,
	identityRepo.InitIdentityQueryRepository,
)

func InitializeContainer(
	appCfg *config.AppConfig,
	readDb *database.GormReadDb,
	writeDb *database.GormWriteDb,
	redisClient redis.Client,
) *api.ApiContainer {
	wire.Build(repoSet, serviceSet, handlerSet, apiSet, container)
	return &api.ApiContainer{}
}

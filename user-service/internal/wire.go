//go:build wireinject
// +build wireinject

package internal

import (
	categoryService "github.com/vucongthanh92/courier/user-service/internal/application/category"
	"github.com/vucongthanh92/courier/user-service/internal/application/cronjob"
	productService "github.com/vucongthanh92/courier/user-service/internal/application/product"
	supplierService "github.com/vucongthanh92/courier/user-service/internal/application/supplier"

	categoryRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/category"
	productRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/product"
	supplierRepo "github.com/vucongthanh92/courier/user-service/internal/repository/persistent/supplier"

	grpcserver "github.com/vucongthanh92/courier/user-service/internal/api/grpc"
	v1 "github.com/vucongthanh92/courier/user-service/internal/api/http/v1"

	"github.com/vucongthanh92/courier/user-service/config"
	"github.com/vucongthanh92/courier/user-service/database"
	"github.com/vucongthanh92/courier/user-service/internal/api"
	"github.com/vucongthanh92/courier/user-service/internal/api/cron"
	"github.com/vucongthanh92/courier/user-service/internal/api/http"
	"github.com/vucongthanh92/courier/user-service/redis"

	"github.com/google/wire"
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
	v1.NewProductHandler,
	v1.NewCategoryHandler,
	v1.NewSupplierHandler,
)

var serviceSet = wire.NewSet(
	cronjob.NewCronJobService,
	productService.NewProductService,
	categoryService.NewCategoryService,
	supplierService.NewSupplierService,
)

var repoSet = wire.NewSet(
	productRepo.NewProductCommandRepository,
	productRepo.NewProductQueryRepository,
	categoryRepo.NewCategoryCommandRepository,
	categoryRepo.NewCategoryQueryRepository,
	supplierRepo.NewSupplierCommandRepository,
	supplierRepo.NewSupplierQueryRepository,
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

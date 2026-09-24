//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/google/wire"
	"github.com/vucongthanh92/courier/payment-gateway/config"
	"github.com/vucongthanh92/courier/payment-gateway/database"
	"github.com/vucongthanh92/courier/payment-gateway/helper/transaction"
	"github.com/vucongthanh92/courier/payment-gateway/internal/api"
	grpcapi "github.com/vucongthanh92/courier/payment-gateway/internal/api/grpc"
	httpapi "github.com/vucongthanh92/courier/payment-gateway/internal/api/http"
	v1 "github.com/vucongthanh92/courier/payment-gateway/internal/api/http/v1"
	providers "github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/providers"
	"github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/providers/sepay"
	redisrepo "github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/redis"
	usergrpc "github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/user_grpc"
	idempotency "github.com/vucongthanh92/courier/payment-gateway/internal/repository/persistent/idempotency"
	ledger "github.com/vucongthanh92/courier/payment-gateway/internal/repository/persistent/ledger"
	outbox "github.com/vucongthanh92/courier/payment-gateway/internal/repository/persistent/outbox"
	providerEvent "github.com/vucongthanh92/courier/payment-gateway/internal/repository/persistent/provider_event"
	topuprepo "github.com/vucongthanh92/courier/payment-gateway/internal/repository/persistent/topup"
	wallet "github.com/vucongthanh92/courier/payment-gateway/internal/repository/persistent/wallet"
	idempotencyuc "github.com/vucongthanh92/courier/payment-gateway/internal/usecase/idempotency"
	topup "github.com/vucongthanh92/courier/payment-gateway/internal/usecase/topup"
	walletuc "github.com/vucongthanh92/courier/payment-gateway/internal/usecase/wallet"
	webhook "github.com/vucongthanh92/courier/payment-gateway/internal/usecase/webhook"
	"github.com/vucongthanh92/courier/payment-gateway/redis"
)

var repoSet = wire.NewSet(
	transaction.InitManagerTxn,
	wallet.InitWalletCmdRepository, wallet.InitWalletQueryRepository,
	topuprepo.InitTopUpCmdRepository, topuprepo.InitTopUpQueryRepository,
	idempotency.InitIdempotencyCmdRepository,
	providerEvent.InitProviderEventCmdRepository,
	ledger.InitLedgerCmdRepository, ledger.InitLedgerQueryRepository,
	outbox.InitOutboxCmdRepository,
	sepay.New, providers.InitPaymentGatewayResolver,
	redisrepo.InitRedisDenylist, redisrepo.InitJWKCacheRepo, usergrpc.NewGrpcClient,
)
var serviceSet = wire.NewSet(
	idempotencyuc.InitIdempotencyUsecase,
	walletuc.InitWalletUsecase,
	topup.InitTopUpUsecase,
	webhook.InitSePayWebhookUsecase,
)
var handlerSet = wire.NewSet(v1.InitTopUpHandler, v1.InitSePayWebhookHandler)
var apiSet = wire.NewSet(httpapi.NewServer, grpcapi.NewServer, api.NewApiContainer)

func provideSePayConfig(cfg *config.AppConfig) config.SePayConfig { return cfg.SePay }

func InitializeContainer(cfg *config.AppConfig, readDB *database.GormReadDb, writeDB *database.GormWriteDb, redisClient redis.Client) *api.ApiContainer {
	wire.Build(repoSet, serviceSet, handlerSet, apiSet, provideSePayConfig)
	return &api.ApiContainer{}
}

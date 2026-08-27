package startup

import (
	"flag"
	"log"

	"github.com/vucongthanh92/courier/payment-gateway/config"
	"github.com/vucongthanh92/courier/payment-gateway/database"
	apihttp "github.com/vucongthanh92/courier/payment-gateway/internal/api/http"
	v1 "github.com/vucongthanh92/courier/payment-gateway/internal/api/http/v1"
	cacheRepo "github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/redis"
	"github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/sepay"
	usergrpc "github.com/vucongthanh92/courier/payment-gateway/internal/repository/external/user_grpc"
	topuprepo "github.com/vucongthanh92/courier/payment-gateway/internal/repository/persistent/topup"
	walletrepo "github.com/vucongthanh92/courier/payment-gateway/internal/repository/persistent/wallet"
	topupuc "github.com/vucongthanh92/courier/payment-gateway/internal/usecase/topup"
	webhookuc "github.com/vucongthanh92/courier/payment-gateway/internal/usecase/webhook"
	redisclient "github.com/vucongthanh92/courier/payment-gateway/redis"
)

func Execute() {
	configPath := flag.String("config", "./config/local/config.yaml", "path to config file")
	flag.Parse()
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Open(cfg.Database.WriteDbCfg.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	redis := redisclient.Open(cfg.Redis)
	provider := sepay.New(cfg.SePay)
	topUpUsecase := topupuc.New(walletrepo.New(db), topuprepo.New(db), provider)
	webhookUsecase := webhookuc.New(db, provider)
	server := apihttp.NewServer(cfg, v1.InitTopUpHandler(topUpUsecase), v1.InitSePayWebhookHandler(provider, webhookUsecase), cacheRepo.InitJWKCacheRepo(redis), usergrpc.NewGrpcClient(cfg), cacheRepo.InitRedisDenylist(redis))
	server.Run()
}

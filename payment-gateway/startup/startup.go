package startup

import (
	"flag"
	"log"

	"github.com/vucongthanh92/courier/payment-gateway/config"
	"github.com/vucongthanh92/courier/payment-gateway/database"
	"github.com/vucongthanh92/courier/payment-gateway/internal"
	"github.com/vucongthanh92/courier/payment-gateway/redis"
)

func Execute() {
	configPath := flag.String("config", "./config/local/config.yaml", "path to config file")
	flag.Parse()
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	readDB, writeDB := database.GetConnectByGorm(cfg.Database)
	if readDB == nil || writeDB == nil {
		log.Fatal("open payment-gateway database connections")
	}
	container := internal.InitializeContainer(cfg, &readDB, &writeDB, redis.Open(cfg.Redis))
	go container.GrpcServer.Run()
	container.HttpServer.Run()
}

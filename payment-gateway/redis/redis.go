package redis

import (
	"github.com/redis/go-redis/v9"
	"github.com/vucongthanh92/courier/payment-gateway/config"
	redisutils "github.com/vucongthanh92/go-base-utils/redis"
)

type Client redis.UniversalClient

func Open(cfg *config.RedisConfig) Client {
	return redisutils.NewUniversalRedisClient(redisutils.Config{Addrs: cfg.Addrs, Password: cfg.Password, Username: cfg.Username, DB: cfg.DB, PoolSize: cfg.PoolSize})
}

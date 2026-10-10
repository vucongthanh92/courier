package redis

import (
	"github.com/redis/go-redis/v9"
	"github.com/vucongthanh92/go-base-utils/models"
	redisUtils "github.com/vucongthanh92/go-base-utils/redis"
)

type Client redis.UniversalClient

func Open(cfg *models.RedisConfig) Client {
	return redisUtils.NewUniversalRedisClient(redisUtils.Config{
		Addrs:    cfg.Addrs,
		Password: cfg.Password,
		Username: cfg.Username,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
}

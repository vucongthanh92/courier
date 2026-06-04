package redis

import (
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

const (
	maxRetries      = 3
	minRetryBackoff = 300 * time.Millisecond
	maxRetryBackoff = 500 * time.Millisecond
	dialTimeout     = 3 * time.Second
	readTimeout     = 3 * time.Second
	writeTimeout    = 3 * time.Second
	minIdleConns    = 20
	poolTimeout     = 6 * time.Second
	idleTimeout     = 12 * time.Second
)

func NewUniversalRedisClient(cfg Config) redis.UniversalClient {
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:                 cfg.Addrs,
		DB:                    cfg.DB,
		Password:              cfg.Password,
		MaxRetries:            maxRetries,
		MinRetryBackoff:       minRetryBackoff,
		MaxRetryBackoff:       maxRetryBackoff,
		DialTimeout:           dialTimeout,
		ReadTimeout:           readTimeout,
		WriteTimeout:          writeTimeout,
		ContextTimeoutEnabled: false,
		PoolFIFO:              false,
		PoolSize:              cfg.PoolSize,
		PoolTimeout:           poolTimeout,
		MinIdleConns:          minIdleConns,
		ConnMaxIdleTime:       idleTimeout,
	})
	_ = redisotel.InstrumentMetrics(rdb)
	_ = redisotel.InstrumentTracing(rdb, redisotel.WithDBStatement(true))
	return rdb
}

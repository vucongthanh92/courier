package utils

import (
	"context"
	"encoding/json"
	"time"

	cacheV9 "github.com/go-redis/cache/v9"
	"github.com/vucongthanh92/go-base-utils/cache"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

// GetQueryCache retrieves a cached value from the provided cache using the specified key
// and unmarshals it into the provided data structure.
func GetQueryCache[T any](cache cache.CacheInterface[string], ctx context.Context, key string, data T) (err error) {
	derivedCtx := context.WithValue(context.Background(), logger.TraceKey, ctx.Value(logger.TraceKey))
	val, err := cache.Get(derivedCtx, key)
	if err != nil {
		if err != cacheV9.ErrCacheMiss {
			logger.WarnCtx(derivedCtx, "GetQueryCache Error", zap.Error(err))
		}
		return err
	}
	return json.Unmarshal([]byte(*val), data)
}

// SetQueryCache stores a value in the provided cache with the specified key and duration,
// after marshaling the data into JSON format.
func SetQueryCache[T any](cache cache.CacheInterface[string], ctx context.Context, key string, duration time.Duration, data T) {
	derivedCtx := context.WithValue(context.Background(), logger.TraceKey, ctx.Value(logger.TraceKey))

	valSave, _ := json.Marshal(data)
	errRedis := cache.Set(derivedCtx, key, string(valSave), duration)
	if errRedis != nil {
		logger.WarnCtx(derivedCtx, "SetQueryCache Error", zap.Error(errRedis))
		return
	}
}

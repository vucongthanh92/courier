package cache

import (
	"context"

	"time"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
)

const (
	defaultCacheSize = 1000
	defaultCacheTTL  = 5 * time.Minute
)

type CacheInterface[T any] interface {
	Get(ctx context.Context, key string) (*T, error)
	Set(ctx context.Context, key string, value T, ttl time.Duration) error
	Del(ctx context.Context, key string) error
}

type cacheImpl[T any] struct {
	cache *cache.Cache
}

func (c cacheImpl[T]) Del(ctx context.Context, key string) (err error) {
	return c.cache.Delete(ctx, key)
}

func (c cacheImpl[T]) Get(ctx context.Context, key string) (val *T, err error) {
	var value T
	if err := c.cache.Get(ctx, key, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

func (c cacheImpl[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) (err error) {
	item := &cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
		TTL:   ttl,
		Do:    nil,
	}
	if err := c.cache.Set(item); err != nil {
		return err
	}
	return nil
}

func NewCache[T any](rdb redis.UniversalClient, size int, ttl time.Duration) CacheInterface[T] {
	if size == 0 {
		size = defaultCacheSize
	}
	if ttl == 0 {
		ttl = defaultCacheTTL
	}
	return &cacheImpl[T]{
		cache: cache.New(&cache.Options{
			Redis:      rdb,
			LocalCache: cache.NewTinyLFU(size, ttl),
		})}
}

func NewInmemoryCache[T any](size int, ttl time.Duration) CacheInterface[T] {
	return NewCache[T](nil, size, ttl)
}

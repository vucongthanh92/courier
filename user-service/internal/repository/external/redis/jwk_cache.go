package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/vucongthanh92/courier/user-service/helper/constants"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
	redisclient "github.com/vucongthanh92/courier/user-service/redis"
)

type JWKCacheRepo interface {
	GetByKid(ctx context.Context, kid string) (*models.JWKCacheEntry, error)
	SetByKid(ctx context.Context, entry models.JWKCacheEntry, ttl time.Duration) error
}

type jwkCacheRepo struct {
	client redisclient.Client
}

func InitJWKCacheRepo(client redisclient.Client) JWKCacheRepo {
	return &jwkCacheRepo{client: client}
}

func (r *jwkCacheRepo) key(kid string) string {
	return fmt.Sprintf("%s%s", constants.JwkCacheKeyPrefix, kid)
}

func (r *jwkCacheRepo) GetByKid(ctx context.Context, kid string) (*models.JWKCacheEntry, error) {
	val, err := r.client.Get(ctx, r.key(kid)).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var entry models.JWKCacheEntry
	if err := json.Unmarshal([]byte(val), &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *jwkCacheRepo) SetByKid(ctx context.Context, entry models.JWKCacheEntry, ttl time.Duration) error {
	buff, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(entry.Kid), buff, ttl).Err()
}

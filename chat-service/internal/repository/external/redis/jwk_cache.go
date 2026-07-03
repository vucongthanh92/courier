package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	redisclient "github.com/vucongthanh92/courier/chat-service/redis"
)

const jwkCacheKeyPrefix = "auth:jwk:kid:"

type JWKCacheEntry struct {
	Kid       string `json:"kid"`
	PublicPEM string `json:"public_pem"`
	Alg       string `json:"alg,omitempty"`
}

type JWKCacheRepo interface {
	GetByKid(ctx context.Context, kid string) (*JWKCacheEntry, error)
	SetByKid(ctx context.Context, entry JWKCacheEntry, ttl time.Duration) error
}

type jwkCacheRepo struct {
	client redisclient.Client
}

func InitJWKCacheRepo(client redisclient.Client) JWKCacheRepo {
	return &jwkCacheRepo{client: client}
}

func (r *jwkCacheRepo) key(kid string) string {
	return fmt.Sprintf("%s%s", jwkCacheKeyPrefix, kid)
}

func (r *jwkCacheRepo) GetByKid(ctx context.Context, kid string) (*JWKCacheEntry, error) {
	val, err := r.client.Get(ctx, r.key(kid)).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var entry JWKCacheEntry
	if err := json.Unmarshal([]byte(val), &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *jwkCacheRepo) SetByKid(ctx context.Context, entry JWKCacheEntry, ttl time.Duration) error {
	buff, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(entry.Kid), buff, ttl).Err()
}

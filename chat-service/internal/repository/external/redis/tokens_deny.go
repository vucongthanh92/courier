package redis

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	redisclient "github.com/vucongthanh92/courier/chat-service/redis"
)

const keyPrefix = "deny:jti:"

type redisDenylist struct {
	redisClient redisclient.Client
}

func InitRedisDenylist(client redisclient.Client) interfaces.TokenDenylistI {
	return &redisDenylist{redisClient: client}
}

func (repo *redisDenylist) Block(ctx context.Context, jti string, ttl time.Duration) error {
	return repo.redisClient.Set(ctx, keyPrefix+jti, "1", ttl).Err()
}

func (repo *redisDenylist) IsBlocked(ctx context.Context, jti string) (bool, error) {
	v, err := repo.redisClient.Exists(ctx, keyPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return v > 0, nil
}

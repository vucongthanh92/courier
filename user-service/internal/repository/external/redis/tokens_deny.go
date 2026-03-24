package redis

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/redis"
)

const keyPrefix = "deny:jti:"

type redisDenylist struct {
	redisClient redis.Client
}

func InitRedisDenylist(client redis.Client) interfaces.TokenDenylistI {
	return &redisDenylist{
		redisClient: client,
	}
}

// Block adds the given jti to the denylist with the specified TTL.
func (repo *redisDenylist) Block(ctx context.Context, jti string, ttl time.Duration) error {
	return repo.redisClient.Set(ctx, keyPrefix+jti, "1", ttl).Err()
}

// IsBlocked checks if the given jti is in the denylist.
func (repo *redisDenylist) IsBlocked(ctx context.Context, jti string) (bool, error) {
	v, err := repo.redisClient.Exists(ctx, keyPrefix+jti).Result()
	if err != nil {
		return false, err
	}

	return v > 0, nil
}

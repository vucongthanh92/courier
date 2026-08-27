package redis

import (
	"context"
	"time"

	"github.com/vucongthanh92/courier/payment-gateway/internal/domain/interfaces"
	redisclient "github.com/vucongthanh92/courier/payment-gateway/redis"
)

type redisDenylist struct{ client redisclient.Client }

func InitRedisDenylist(client redisclient.Client) interfaces.TokenDenylistI {
	return &redisDenylist{client: client}
}
func (r *redisDenylist) Block(ctx context.Context, jti string, ttl time.Duration) error {
	return r.client.Set(ctx, "payment:deny:jti:"+jti, "1", ttl).Err()
}
func (r *redisDenylist) IsBlocked(ctx context.Context, jti string) (bool, error) {
	exists, err := r.client.Exists(ctx, "payment:deny:jti:"+jti).Result()
	return exists > 0, err
}

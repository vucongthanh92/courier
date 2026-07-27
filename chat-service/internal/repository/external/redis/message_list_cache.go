package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	redisclient "github.com/vucongthanh92/courier/chat-service/redis"
)

type messageListCache struct {
	client redisclient.Client
}

func InitMessageListCache(client redisclient.Client) interfaces.MessageListCacheI {
	return &messageListCache{client: client}
}

func (r *messageListCache) latestKey(conversationID uint64, limit int) string {
	return fmt.Sprintf("chat:conversation:%d:messages:latest:%d", conversationID, limit)
}

func (r *messageListCache) latestPattern(conversationID uint64) string {
	return fmt.Sprintf("chat:conversation:%d:messages:latest:*", conversationID)
}

func (r *messageListCache) GetLatest(ctx context.Context, conversationID uint64, limit int) (*models.CachedMessageListPage, error) {
	val, err := r.client.Get(ctx, r.latestKey(conversationID, limit)).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var page models.CachedMessageListPage
	if err := json.Unmarshal([]byte(val), &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *messageListCache) SetLatest(ctx context.Context, conversationID uint64, limit int, page models.CachedMessageListPage, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = constants.Time_Cache_15_minutes
	}
	buff, err := json.Marshal(page)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.latestKey(conversationID, limit), buff, ttl).Err()
}

func (r *messageListCache) InvalidateLatest(ctx context.Context, conversationID uint64) error {
	iter := r.client.Scan(ctx, 0, r.latestPattern(conversationID), 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

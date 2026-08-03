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

type userProfileCache struct {
	client redisclient.Client
}

func InitUserProfileCache(client redisclient.Client) interfaces.UserProfileCacheI {
	return &userProfileCache{client: client}
}

func (r *userProfileCache) key(userID uint64) string {
	return fmt.Sprintf("chat:user_profile:%d", userID)
}

func (r *userProfileCache) GetMany(ctx context.Context, userIDs []uint64) (map[uint64]models.UserProfileSummaryResponse, error) {
	profiles := make(map[uint64]models.UserProfileSummaryResponse, len(userIDs))
	if len(userIDs) == 0 {
		return profiles, nil
	}

	keys := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		keys = append(keys, r.key(userID))
	}

	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	for _, value := range values {
		if value == nil {
			continue
		}
		raw, ok := value.(string)
		if !ok {
			continue
		}
		var profile models.UserProfileSummaryResponse
		if err := json.Unmarshal([]byte(raw), &profile); err != nil {
			return nil, err
		}
		if profile.UserID != 0 {
			profiles[profile.UserID] = profile
		}
	}

	return profiles, nil
}

func (r *userProfileCache) SetMany(ctx context.Context, profiles []models.UserProfileSummaryResponse, ttl time.Duration) error {
	if len(profiles) == 0 {
		return nil
	}
	if ttl <= 0 {
		ttl = constants.Time_Cache_5_minutes
	}

	pipe := r.client.Pipeline()
	for _, profile := range profiles {
		if profile.UserID == 0 {
			continue
		}
		buff, err := json.Marshal(profile)
		if err != nil {
			return err
		}
		pipe.Set(ctx, r.key(profile.UserID), buff, ttl)
	}
	_, err := pipe.Exec(ctx)
	if err == goredis.Nil {
		return nil
	}
	return err
}

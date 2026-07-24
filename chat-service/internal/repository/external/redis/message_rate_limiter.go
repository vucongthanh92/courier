package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/vucongthanh92/courier/chat-service/config"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	redisclient "github.com/vucongthanh92/courier/chat-service/redis"
)

var messageRateLimitScript = goredis.NewScript(`
local limits = {tonumber(ARGV[1]), tonumber(ARGV[3]), tonumber(ARGV[5])}
local windows = {tonumber(ARGV[2]), tonumber(ARGV[4]), tonumber(ARGV[6])}
local retry_after = 0

for i = 1, 3 do
  local current = tonumber(redis.call("GET", KEYS[i]) or "0")
  if current >= limits[i] then
    local ttl = redis.call("PTTL", KEYS[i])
    if ttl < 0 then
      ttl = windows[i]
    end
    retry_after = math.max(retry_after, math.ceil(ttl / 1000))
  end
end

if retry_after > 0 then
  return {0, retry_after, 0}
end

local remaining = limits[1]
for i = 1, 3 do
  local current = redis.call("INCR", KEYS[i])
  if current == 1 then
    redis.call("PEXPIRE", KEYS[i], windows[i])
  end
  remaining = math.min(remaining, limits[i] - current)
end

return {1, 0, remaining}
`)

type messageRateLimiter struct {
	client redisclient.Client
	cfg    config.MessageRateLimitConfig
}

func InitMessageRateLimiter(client redisclient.Client, appCfg *config.AppConfig) interfaces.MessageRateLimiterI {
	cfg := defaultMessageRateLimitConfig()
	if appCfg.MessageRateLimit != nil {
		cfg = *appCfg.MessageRateLimit
		applyMessageRateLimitDefaults(&cfg)
	}
	return &messageRateLimiter{client: client, cfg: cfg}
}

func (r *messageRateLimiter) Allow(ctx context.Context, userID, conversationID uint64) (interfaces.MessageRateLimitResult, error) {
	if !r.cfg.Enabled {
		return interfaces.MessageRateLimitResult{Allowed: true}, nil
	}

	keys := []string{
		fmt.Sprintf("chat:message:rate:{%d}:burst:%d", userID, conversationID),
		fmt.Sprintf("chat:message:rate:{%d}:conversation:%d", userID, conversationID),
		fmt.Sprintf("chat:message:rate:{%d}:user", userID),
	}
	args := []any{
		r.cfg.Burst.Limit, r.cfg.Burst.WindowSeconds * 1000,
		r.cfg.Conversation.Limit, r.cfg.Conversation.WindowSeconds * 1000,
		r.cfg.User.Limit, r.cfg.User.WindowSeconds * 1000,
	}

	values, err := messageRateLimitScript.Run(ctx, r.client, keys, args...).Int64Slice()
	if err != nil {
		return interfaces.MessageRateLimitResult{}, err
	}
	if len(values) != 3 {
		return interfaces.MessageRateLimitResult{}, fmt.Errorf("unexpected message rate limiter response length: %d", len(values))
	}

	retryAfter := int(values[1])
	result := interfaces.MessageRateLimitResult{
		Allowed:    values[0] == 1,
		Limit:      r.cfg.Burst.Limit,
		Remaining:  int(values[2]),
		RetryAfter: retryAfter,
	}
	if retryAfter > 0 {
		result.ResetAt = time.Now().Add(time.Duration(retryAfter) * time.Second).Unix()
	}
	return result, nil
}

func defaultMessageRateLimitConfig() config.MessageRateLimitConfig {
	return config.MessageRateLimitConfig{
		Enabled:      true,
		Burst:        config.RateLimitWindowConfig{Limit: 5, WindowSeconds: 2},
		Conversation: config.RateLimitWindowConfig{Limit: 30, WindowSeconds: 60},
		User:         config.RateLimitWindowConfig{Limit: 100, WindowSeconds: 60},
	}
}

func applyMessageRateLimitDefaults(cfg *config.MessageRateLimitConfig) {
	defaults := defaultMessageRateLimitConfig()
	if cfg.Burst.Limit <= 0 {
		cfg.Burst.Limit = defaults.Burst.Limit
	}
	if cfg.Burst.WindowSeconds <= 0 {
		cfg.Burst.WindowSeconds = defaults.Burst.WindowSeconds
	}
	if cfg.Conversation.Limit <= 0 {
		cfg.Conversation.Limit = defaults.Conversation.Limit
	}
	if cfg.Conversation.WindowSeconds <= 0 {
		cfg.Conversation.WindowSeconds = defaults.Conversation.WindowSeconds
	}
	if cfg.User.Limit <= 0 {
		cfg.User.Limit = defaults.User.Limit
	}
	if cfg.User.WindowSeconds <= 0 {
		cfg.User.WindowSeconds = defaults.User.WindowSeconds
	}
}

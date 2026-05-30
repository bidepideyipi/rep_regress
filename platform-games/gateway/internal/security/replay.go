package security

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ReplayProtection struct {
	redis    *redis.Client
	ttl      time.Duration
	enabled  bool
}

func NewReplayProtection(redisClient *redis.Client, ttl time.Duration, enabled bool) *ReplayProtection {
	return &ReplayProtection{
		redis:   redisClient,
		ttl:     ttl,
		enabled: enabled,
	}
}

func (rp *ReplayProtection) IsReplayRequest(ctx context.Context, requestID string) (bool, error) {
	if !rp.enabled {
		return false, nil
	}

	key := fmt.Sprintf("replay:%s", requestID)

	exists, err := rp.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check request ID: %w", err)
	}

	if exists > 0 {
		return true, nil
	}

	if err := rp.redis.Set(ctx, key, "1", rp.ttl).Err(); err != nil {
		return false, fmt.Errorf("failed to store request ID: %w", err)
	}

	return false, nil
}

func (rp *ReplayProtection) MarkRequest(ctx context.Context, requestID string) error {
	if !rp.enabled {
		return nil
	}

	key := fmt.Sprintf("replay:%s", requestID)
	return rp.redis.Set(ctx, key, "1", rp.ttl).Err()
}

func (rp *ReplayProtection) Cleanup(ctx context.Context) error {
	if !rp.enabled {
		return nil
	}

	iter := rp.redis.Scan(ctx, 0, "replay:*", 100).Iterator()
	keys := make([]string, 0)

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		if err := rp.redis.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

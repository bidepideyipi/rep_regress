package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/platform-games/gateway/internal/models"
)

type Cache struct {
	redis *redis.Client
}

func NewCache(redisClient *redis.Client) *Cache {
	return &Cache{redis: redisClient}
}

func (c *Cache) GetMerchant(ctx context.Context, merchantID string) (*models.MerchantCache, error) {
	key := fmt.Sprintf("merchant:%s", merchantID)
	data, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get merchant from cache: %w", err)
	}

	var merchant models.MerchantCache
	if err := json.Unmarshal([]byte(data), &merchant); err != nil {
		return nil, fmt.Errorf("failed to unmarshal merchant: %w", err)
	}

	return &merchant, nil
}

func (c *Cache) SetMerchant(ctx context.Context, merchant *models.MerchantCache, ttl time.Duration) error {
	key := fmt.Sprintf("merchant:%s", merchant.MerchantID)
	data, err := json.Marshal(merchant)
	if err != nil {
		return fmt.Errorf("failed to marshal merchant: %w", err)
	}

	return c.redis.Set(ctx, key, data, ttl).Err()
}

func (c *Cache) GetUser(ctx context.Context, userID string) (*models.UserCache, error) {
	key := fmt.Sprintf("user:%s", userID)
	data, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user from cache: %w", err)
	}

	var user models.UserCache
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

func (c *Cache) SetUser(ctx context.Context, user *models.UserCache, ttl time.Duration) error {
	key := fmt.Sprintf("user:%s", user.UserID)
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	return c.redis.Set(ctx, key, data, ttl).Err()
}

func (c *Cache) GetGame(ctx context.Context, gameID string) (*models.Game, error) {
	key := fmt.Sprintf("game:%s", gameID)
	data, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get game from cache: %w", err)
	}

	var game models.Game
	if err := json.Unmarshal([]byte(data), &game); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game: %w", err)
	}

	return &game, nil
}

func (c *Cache) SetGame(ctx context.Context, game *models.Game, ttl time.Duration) error {
	key := fmt.Sprintf("game:%s", game.GameID)
	data, err := json.Marshal(game)
	if err != nil {
		return fmt.Errorf("failed to marshal game: %w", err)
	}

	return c.redis.Set(ctx, key, data, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.redis.Del(ctx, key).Err()
}

func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return n > 0, nil
}

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// UserRTPData 用户RTP缓存数据
type UserRTPData struct {
	TotalBet   float64   `json:"total_bet"`
	TotalWin   float64   `json:"total_win"`
	NetResult  float64   `json:"net_result"`
	RTP        float64   `json:"rtp"`
	TotalSpins uint64    `json:"total_spins"`
	AvgBet     float64   `json:"avg_bet"`
	FirstSpin  time.Time `json:"first_spin"`
	LastSpin   time.Time `json:"last_spin"`
}

// RedisCache Redis缓存客户端
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache 创建Redis缓存客户端
func NewRedisCache(config *RedisConfig) (*RedisCache, error) {
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}

	log.Printf("Redis连接成功: %s:%d (DB: %d)", config.Host, config.Port, config.DB)

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// GetUserRTP 获取用户RTP缓存
func (rc *RedisCache) GetUserRTP(userID string) (*UserRTPData, error) {
	key := rc.userRTPKey(userID)

	data, err := rc.client.Get(rc.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // 缓存不存在
	}
	if err != nil {
		return nil, fmt.Errorf("获取缓存失败: %w", err)
	}

	var rtpData UserRTPData
	if err := json.Unmarshal([]byte(data), &rtpData); err != nil {
		return nil, fmt.Errorf("解析缓存数据失败: %w", err)
	}

	return &rtpData, nil
}

// SetUserRTP 设置用户RTP缓存
func (rc *RedisCache) SetUserRTP(userID string, data *UserRTPData, ttl time.Duration) error {
	key := rc.userRTPKey(userID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	if err := rc.client.Set(rc.ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("设置缓存失败: %w", err)
	}

	return nil
}

// userRTPKey 生成用户RTP缓存键
func (rc *RedisCache) userRTPKey(userID string) string {
	return fmt.Sprintf("rtp:user:%s", userID)
}

// DeleteUserRTP 删除用户RTP缓存
func (rc *RedisCache) DeleteUserRTP(userID string) error {
	key := rc.userRTPKey(userID)
	if err := rc.client.Del(rc.ctx, key).Err(); err != nil {
		return fmt.Errorf("删除缓存失败: %w", err)
	}
	return nil
}

// Close 关闭连接
func (rc *RedisCache) Close() error {
	if rc.client != nil {
		return rc.client.Close()
	}
	return nil
}

// GetStats 获取Redis统计信息
func (rc *RedisCache) GetStats() (map[string]string, error) {
	info, err := rc.client.Info(rc.ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"info": info,
	}, nil
}

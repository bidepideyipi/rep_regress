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

// freeSpinKey 生成Free Spin缓存键
// 格式: freespin:{integrator_id}:{user_id}:{game_id}
func (rc *RedisCache) freeSpinKey(integratorID, userID, gameID string) string {
	return fmt.Sprintf("freespin:%s:%s:%s", integratorID, userID, gameID)
}

// GetFreeSpinRemaining 获取剩余Free Spin次数
func (rc *RedisCache) GetFreeSpinRemaining(integratorID, userID, gameID string) (int64, error) {
	key := rc.freeSpinKey(integratorID, userID, gameID)

	result, err := rc.client.Get(rc.ctx, key).Result()
	if err == redis.Nil {
		return 0, nil // 不存在则表示没有free spin
	}
	if err != nil {
		return 0, fmt.Errorf("获取Free Spin失败: %w", err)
	}

	var remaining int64
	if _, err := fmt.Sscanf(result, "%d", &remaining); err != nil {
		return 0, fmt.Errorf("解析Free Spin数据失败: %w", err)
	}

	return remaining, nil
}

// SetFreeSpinRemaining 设置剩余Free Spin次数
func (rc *RedisCache) SetFreeSpinRemaining(integratorID, userID, gameID string, remaining int64, ttl time.Duration) error {
	key := rc.freeSpinKey(integratorID, userID, gameID)

	if err := rc.client.Set(rc.ctx, key, remaining, ttl).Err(); err != nil {
		return fmt.Errorf("设置Free Spin失败: %w", err)
	}

	return nil
}

// AddFreeSpinRemaining 增加剩余Free Spin次数（触发新free spin时使用）
func (rc *RedisCache) AddFreeSpinRemaining(integratorID, userID, gameID string, addCount int64, ttl time.Duration) (int64, error) {
	key := rc.freeSpinKey(integratorID, userID, gameID)

	// 使用原子操作增加次数，并设置过期时间
	pipe := rc.client.Pipeline()
	incrCmd := pipe.IncrBy(rc.ctx, key, addCount)
	pipe.Expire(rc.ctx, key, ttl)

	if _, err := pipe.Exec(rc.ctx); err != nil {
		return 0, fmt.Errorf("增加Free Spin失败: %w", err)
	}

	return incrCmd.Val(), nil
}

// DecrementFreeSpin 扣减一次Free Spin（原子操作）
func (rc *RedisCache) DecrementFreeSpin(integratorID, userID, gameID string) (int64, error) {
	key := rc.freeSpinKey(integratorID, userID, gameID)

	// 使用原子递减操作
	result, err := rc.client.Decr(rc.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("扣减Free Spin失败: %w", err)
	}

	// 如果扣减后为0或负数，删除key
	if result <= 0 {
		rc.client.Del(rc.ctx, key)
		return 0, nil
	}

	return result, nil
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

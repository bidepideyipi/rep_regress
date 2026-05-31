package rtp

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/shopspring/decimal"
	"platform-games/slot-game/cache"
)

// UserRTPService 用户RTP服务（带缓存）
type UserRTPService struct {
	ckService   *RTPService
	redisCache  *cache.RedisCache
	cacheExpiry time.Duration
}

// NewUserRTPService 创建用户RTP服务
func NewUserRTPService(ckHost string, ckPort int, ckUsername, ckPassword, ckDatabase string, redisCache *cache.RedisCache, cacheExpiry time.Duration) (*UserRTPService, error) {
	ckService, err := NewRTPService(ckHost, ckPort, ckUsername, ckPassword, ckDatabase)
	if err != nil {
		return nil, err
	}

	return &UserRTPService{
		ckService:   ckService,
		redisCache:  redisCache,
		cacheExpiry: cacheExpiry,
	}, nil
}

// UserRTPSummary 用户RTP汇总数据
type UserRTPSummary struct {
	TotalBet   float64   `json:"total_bet"`
	TotalWin   float64   `json:"total_win"`
	NetResult  float64   `json:"net_result"`
	RTP        float64   `json:"rtp"`
	TotalSpins uint64    `json:"total_spins"`
	AvgBet     float64   `json:"avg_bet"`
	FirstSpin  time.Time `json:"first_spin"`
	LastSpin   time.Time `json:"last_spin"`
}

// GetUserRTP 获取用户RTP（先查缓存，不存在则查ClickHouse）
func (s *UserRTPService) GetUserRTP(ctx context.Context, userID string) (*UserRTPSummary, error) {
	// 先查Redis缓存
	if s.redisCache != nil {
		cachedData, err := s.redisCache.GetUserRTP(userID)
		if err == nil && cachedData != nil {
			log.Printf("缓存命中: user_id=%s", userID)
			return &UserRTPSummary{
				TotalBet:   cachedData.TotalBet,
				TotalWin:   cachedData.TotalWin,
				NetResult:  cachedData.NetResult,
				RTP:        cachedData.RTP,
				TotalSpins: cachedData.TotalSpins,
				AvgBet:     cachedData.AvgBet,
				FirstSpin:  cachedData.FirstSpin,
				LastSpin:   cachedData.LastSpin,
			}, nil
		}
		if err != nil {
			log.Printf("缓存查询失败: %v", err)
		}
	}

	// 缓存不存在，查询ClickHouse
	log.Printf("查询ClickHouse: user_id=%s", userID)
	summary, err := s.getUserRTPFromCK(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	if s.redisCache != nil && summary != nil {
		cacheData := &cache.UserRTPData{
			TotalBet:   summary.TotalBet,
			TotalWin:   summary.TotalWin,
			NetResult:  summary.NetResult,
			RTP:        summary.RTP,
			TotalSpins: summary.TotalSpins,
			AvgBet:     summary.AvgBet,
			FirstSpin:  summary.FirstSpin,
			LastSpin:   summary.LastSpin,
		}
		if err := s.redisCache.SetUserRTP(userID, cacheData, s.cacheExpiry); err != nil {
			log.Printf("写入缓存失败: %v", err)
		} else {
			log.Printf("写入缓存成功: user_id=%s, ttl=%v", userID, s.cacheExpiry)
		}
	}

	return summary, nil
}

// getUserRTPFromCK 从ClickHouse查询用户RTP
func (s *UserRTPService) getUserRTPFromCK(ctx context.Context, userID string) (*UserRTPSummary, error) {
	query := `
		SELECT
			total_bet,
			total_win,
			net_result,
			rtp,
			total_spins,
			avg_bet,
			first_spin,
			last_spin
		FROM (
			SELECT
				sum(total_bet) as total_bet,
				sum(total_win) as total_win,
				sum(net_result) as net_result,
				if(total_bet > 0, round(total_win / total_bet, 4), 0) as rtp,
				sum(spin_count) as total_spins,
				if(total_spins > 0, round(total_bet / total_spins, 2), 0) as avg_bet,
				min(time_window) as first_spin,
				max(time_window) as last_spin
			FROM rtp_analytics.rtp_user_realtime_local
			WHERE user_id = @user_id
		)`

	var (
		totalBet   decimal.Decimal
		totalWin   decimal.Decimal
		netResult  decimal.Decimal
		rtp        decimal.Decimal
		totalSpins uint64
		avgBet     decimal.Decimal
		firstSpin  time.Time
		lastSpin   time.Time
	)

	err := s.ckService.conn.QueryRow(ctx, query,
		clickhouse.Named("user_id", userID),
	).Scan(
		&totalBet,
		&totalWin,
		&netResult,
		&rtp,
		&totalSpins,
		&avgBet,
		&firstSpin,
		&lastSpin,
	)

	if err != nil {
		return nil, fmt.Errorf("查询ClickHouse失败: %w", err)
	}

	summary := &UserRTPSummary{
		TotalBet:   totalBet.InexactFloat64(),
		TotalWin:   totalWin.InexactFloat64(),
		NetResult:  netResult.InexactFloat64(),
		RTP:        rtp.InexactFloat64(),
		TotalSpins: totalSpins,
		AvgBet:     avgBet.InexactFloat64(),
		FirstSpin:  firstSpin,
		LastSpin:   lastSpin,
	}

	return summary, nil
}

// InvalidateUserRTP 使缓存失效
func (s *UserRTPService) InvalidateUserRTP(userID string) error {
	if s.redisCache != nil {
		return s.redisCache.DeleteUserRTP(userID)
	}
	return nil
}

// Close 关闭连接
func (s *UserRTPService) Close() error {
	if s.ckService != nil {
		return s.ckService.Close()
	}
	return nil
}

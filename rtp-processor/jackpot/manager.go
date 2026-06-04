package jackpot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/rtp-processor/db"
	"github.com/shopspring/decimal"
)

// Manager Jackpot管理器
type Manager struct {
	db               *sql.DB
	configCache      map[string]*Config
	configCacheMutex sync.RWMutex
	cacheExpiry      time.Time
	updateBuffer     map[string]map[string]float64
	updateMutex      sync.Mutex
	flushInterval    time.Duration
	lastFlush        time.Time
	flushChan        chan struct{}
}

// Config Jackpot配置
type Config struct {
	GameID       string
	Enabled      bool
	Contribution float64
	MiniRatio    float64
	MinorRatio   float64
	MajorRatio   float64
	GrandRatio   float64
}

// NewManager 创建Jackpot管理器
func NewManager(mysqlMgr *db.MySQLManager) (*Manager, error) {
	if mysqlMgr == nil {
		return nil, fmt.Errorf("MySQL管理器不能为空")
	}

	jm := &Manager{
		db:            mysqlMgr.GetDB(),
		configCache:   make(map[string]*Config),
		updateBuffer:  make(map[string]map[string]float64),
		flushInterval: 5 * time.Second,
		cacheExpiry:   time.Time{},
		flushChan:     make(chan struct{}, 1),
	}

	// 预加载配置
	if err := jm.loadConfig(); err != nil {
		log.Printf("[Jackpot] 预加载配置失败: %v", err)
	}

	log.Printf("[Jackpot] 管理器初始化完成")
	return jm, nil
}

// BetLog 下注日志接口
type BetLog interface {
	GetGameID() string
	GetBetAmount() decimal.Decimal
}

// CalculateJackpot 计算Jackpot金额并加入更新队列
func (jm *Manager) CalculateJackpot(betLog BetLog) {
	betAmount := betLog.GetBetAmount()
	if betAmount.LessThanOrEqual(decimal.Zero) {
		log.Printf("[Jackpot] 跳过: 投注金额<=0, gameID=%s, amount=%s", betLog.GetGameID(), betAmount.String())
		return
	}

	// 获取游戏配置
	cfg := jm.getConfig(betLog.GetGameID())

	var gameID string
	var miniRatio, minorRatio, majorRatio, grandRatio float64

	if cfg != nil && cfg.Enabled {
		gameID = cfg.GameID
		miniRatio = cfg.Contribution * cfg.MiniRatio
		minorRatio = cfg.Contribution * cfg.MinorRatio
		majorRatio = cfg.Contribution * cfg.MajorRatio
		grandRatio = cfg.Contribution * cfg.GrandRatio
	} else {
		log.Printf("[Jackpot] 游戏无专属配置, gameID=%s, 使用全局配置", betLog.GetGameID())
		gameID = "0"
		miniRatio = 0
		minorRatio = 0
		majorRatio = 0

		jm.configCacheMutex.RLock()
		globalCfg := jm.configCache["0"]
		jm.configCacheMutex.RUnlock()

		if globalCfg != nil && globalCfg.Enabled {
			grandRatio = globalCfg.Contribution * globalCfg.GrandRatio
		} else {
			grandRatio = 0.01 * 0.2
		}
	}

	betAmountFloat, _ := betAmount.Float64()

	updateCount := 0
	if miniRatio > 0 {
		jm.AddUpdate(gameID, "mini", betAmountFloat*miniRatio)
		updateCount++
	}
	if minorRatio > 0 {
		jm.AddUpdate(gameID, "minor", betAmountFloat*minorRatio)
		updateCount++
	}
	if majorRatio > 0 {
		jm.AddUpdate(gameID, "major", betAmountFloat*majorRatio)
		updateCount++
	}
	if grandRatio > 0 {
		jm.AddUpdate(gameID, "grand", betAmountFloat*grandRatio)
		updateCount++
	}

	if updateCount == 0 {
		log.Printf("[Jackpot] 跳过: 所有ratio=0, gameID=%s", betLog.GetGameID())
	}
}

// loadConfig 加载Jackpot配置
func (jm *Manager) loadConfig() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT game_id, enabled, contribution_rate, mini_ratio, minor_ratio, major_ratio, grand_ratio
		FROM jackpot_config
		WHERE enabled = 1
	`

	rows, err := jm.db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	jm.configCacheMutex.Lock()
	defer jm.configCacheMutex.Unlock()

	for k := range jm.configCache {
		delete(jm.configCache, k)
	}

	for rows.Next() {
		var (
			gameID       string
			enabled      int8
			contribution float64
			miniRatio    float64
			minorRatio   float64
			majorRatio   float64
			grandRatio   float64
		)

		if err := rows.Scan(&gameID, &enabled, &contribution, &miniRatio, &minorRatio, &majorRatio, &grandRatio); err != nil {
			log.Printf("[Jackpot] 扫描配置失败: %v", err)
			continue
		}

		jm.configCache[gameID] = &Config{
			GameID:       gameID,
			Enabled:      enabled == 1,
			Contribution: contribution * 0.01,
			MiniRatio:    miniRatio,
			MinorRatio:   minorRatio,
			MajorRatio:   majorRatio,
			GrandRatio:   grandRatio,
		}
	}

	jm.cacheExpiry = time.Now().Add(5 * time.Minute)
	log.Printf("[Jackpot] 配置加载完成，共 %d 个游戏", len(jm.configCache))

	return nil
}

// getConfig 获取游戏Jackpot配置
func (jm *Manager) getConfig(gameID string) *Config {
	jm.configCacheMutex.RLock()

	if time.Now().After(jm.cacheExpiry) {
		jm.configCacheMutex.RUnlock()
		if err := jm.loadConfig(); err != nil {
			log.Printf("[Jackpot] 刷新配置失败: %v", err)
		}
		jm.configCacheMutex.RLock()
	}

	config, exists := jm.configCache[gameID]
	jm.configCacheMutex.RUnlock()

	if exists {
		return config
	}
	return nil
}

// AddUpdate 添加池更新
func (jm *Manager) AddUpdate(gameID, poolType string, amount float64) {
	jm.updateMutex.Lock()
	defer jm.updateMutex.Unlock()

	if jm.updateBuffer[gameID] == nil {
		jm.updateBuffer[gameID] = make(map[string]float64)
	}
	jm.updateBuffer[gameID][poolType] += amount

	// 触发异步刷新（非阻塞，防止重复触发）
	select {
	case jm.flushChan <- struct{}{}:
		go jm.Flush()
	default:
		// 已经有 flush 在等待或运行
	}
}

// Flush 刷新Jackpot池更新
func (jm *Manager) Flush() error {
	// 从 channel 确认触发
	select {
	case <-jm.flushChan:
	default:
	}

	// 检查是否需要刷新
	if time.Since(jm.lastFlush) < jm.flushInterval {
		return nil
	}

	jm.updateMutex.Lock()
	if len(jm.updateBuffer) == 0 {
		jm.updateMutex.Unlock()
		return nil
	}

	updates := make(map[string]map[string]float64)
	for gameID, pools := range jm.updateBuffer {
		updates[gameID] = make(map[string]float64)
		for poolType, amount := range pools {
			updates[gameID][poolType] = amount
		}
	}
	jm.updateBuffer = make(map[string]map[string]float64)
	jm.lastFlush = time.Now()
	jm.updateMutex.Unlock()

	return jm.batchUpdatePools(updates)
}

// batchUpdatePools 批量更新Jackpot池
func (jm *Manager) batchUpdatePools(updates map[string]map[string]float64) error {
	if len(updates) == 0 {
		log.Printf("[Jackpot] 无更新数据")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// log.Printf("[Jackpot] 批量更新 %d 个游戏的池子", len(updates))

	updateQuery := `
		INSERT INTO jackpot_pool (game_id, pool_type, current_amount, win_count)
		VALUES (?, ?, ?, 0)
		ON DUPLICATE KEY UPDATE
			current_amount = current_amount + ?,
			win_count = win_count
	`

	updateCount := 0
	for gameID, pools := range updates {
		for poolType, amount := range pools {
			roundedAmount := math.Round(amount*100) / 100
			if roundedAmount < 0.01 && roundedAmount > 0 {
				roundedAmount = 0.01
			}

			result, err := jm.db.ExecContext(ctx, updateQuery, gameID, poolType, roundedAmount, roundedAmount)
			if err != nil {
				log.Printf("[Jackpot] 更新池失败 [%s/%s]: %v", gameID, poolType, err)
				continue
			}

			rowsAffected, _ := result.RowsAffected()
			updateCount++
			if rowsAffected == 1 {
				log.Printf("[Jackpot] 创建新池 [%s/%s]，初始金额: %.2f", gameID, poolType, roundedAmount)
			}
		}
	}

	log.Printf("[Jackpot] 批量更新完成，共更新 %d 个池子", updateCount)
	return nil
}

// Close 关闭Jackpot管理器
func (jm *Manager) Close() error {
	log.Printf("[Jackpot] 关闭前刷新剩余数据")
	if err := jm.Flush(); err != nil {
		log.Printf("[Jackpot] 刷新失败: %v", err)
	}
	return nil
}

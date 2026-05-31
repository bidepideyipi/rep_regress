package models

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/shopspring/decimal"
)

// GameLogDetail 游戏日志详情 - 对应rtp_analytics.game_log_detail_local表
type GameLogDetail struct {
	LogID           string              `ch:"log_id"`
	GameSessionID   string              `ch:"game_session_id"`
	IntegratorID    string              `ch:"integrator_id"`
	UserID          string              `ch:"user_id"`
	GameID          string              `ch:"game_id"`
	BetAmount       decimal.Decimal     `ch:"bet_amount"`
	WinAmount       decimal.Decimal     `ch:"win_amount"`
	NetResult       decimal.Decimal     `ch:"net_result"`
	BetLines        uint16              `ch:"bet_lines"`
	BetPerLine      decimal.Decimal     `ch:"bet_per_line"`
	IsFreeSpin      uint8               `ch:"is_free_spin"`
	BonusFeature    string              `ch:"bonus_feature"`
	DeviceType      string              `ch:"device_type"`
	DeviceOS        string              `ch:"device_os"`
	BrowserType     string              `ch:"browser_type"`
	IPAddress       uint32              `ch:"ip_address"`
	IPRegion        string              `ch:"ip_region"`
	IPCountry       string              `ch:"ip_country"`
	SessionID       string              `ch:"session_id"`
	ServerID        string              `ch:"server_id"`
	ProcessingTime  uint32              `ch:"processing_time_ms"`
	ErrorCode       uint16              `ch:"error_code"`
	ErrorMessage    string              `ch:"error_message"`
	GameResultJSON  string              `ch:"game_result_json"`
	ReelResult      [][]string          `ch:"reel_result"`
	WinLines        [][]interface{}     `ch:"win_lines"` // Tuple(UInt16, Decimal(18,4), Array(String), UInt8)
	UserAgent       string              `ch:"user_agent"`
	LogTime         time.Time           `ch:"log_time"`
}

// WinLineCH ClickHouse中奖线路格式（用于转换）
type WinLineCH struct {
	LineID    uint16
	WinAmount decimal.Decimal
	Symbols   []string
	IsWild    uint8
}

// ToTuple 转换为ClickHouse Tuple格式
func (w *WinLineCH) ToTuple() []interface{} {
	return []interface{}{w.LineID, w.WinAmount, w.Symbols, w.IsWild}
}

// BatchConfig 批量写入配置
type BatchConfig struct {
	BatchSize     int           // 批量大小
	FlushInterval time.Duration // 刷新间隔
	Timeout       time.Duration // 写入超时
	RetryTimes    int           // 重试次数
	RetryInterval time.Duration // 重试间隔
}

// DefaultBatchConfig 默认批量配置
func DefaultBatchConfig() *BatchConfig {
	return &BatchConfig{
		BatchSize:     100,
		FlushInterval: 5 * time.Second,
		Timeout:       10 * time.Second,
		RetryTimes:    3,
		RetryInterval: 100 * time.Millisecond,
	}
}

// BatchWriter 批量写入器
type BatchWriter struct {
	conn        driver.Conn
	config      *BatchConfig
	buffer      []GameLogDetail
	mu          sync.Mutex
	flushTicker *time.Ticker
	done        chan struct{}
	serverIP    string
	tableName   string
}

// NewBatchWriter 创建批量写入器
func NewBatchWriter(host string, port int, username, password, database string, serverIP string, config *BatchConfig) (*BatchWriter, error) {
	if config == nil {
		config = DefaultBatchConfig()
	}

	log.Printf("[ClickHouse] 正在连接 %s:%d, user=%s, database=%s", host, port, username, database)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn := fmt.Sprintf("%s:%d", host, port)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dsn},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("连接ClickHouse失败: %w", err)
	}

	// 测试连接
	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ClickHouse ping失败: %w", err)
	}

	log.Printf("[ClickHouse] 连接成功")

	bw := &BatchWriter{
		conn:        conn,
		config:      config,
		buffer:      make([]GameLogDetail, 0, config.BatchSize),
		flushTicker: time.NewTicker(config.FlushInterval),
		done:        make(chan struct{}),
		serverIP:    serverIP,
		tableName:   "game_log_detail_local",
	}

	// 启动后台刷新协程
	go bw.flushLoop()

	return bw, nil
}

// Write 写入单条日志
func (bw *BatchWriter) Write(entry GameLogDetail) error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	bw.buffer = append(bw.buffer, entry)

	log.Printf("[ClickHouse] 写入缓冲区，当前大小: %d/%d", len(bw.buffer), bw.config.BatchSize)

	// 达到批量大小，触发刷新
	if len(bw.buffer) >= bw.config.BatchSize {
		go func() {
			if err := bw.flush(); err != nil {
				log.Printf("[ClickHouse] 批量刷新失败: %v", err)
			}
		}()
	}

	return nil
}

// flushLoop 定期刷新缓冲区
func (bw *BatchWriter) flushLoop() {
	for {
		select {
		case <-bw.flushTicker.C:
			if err := bw.flush(); err != nil {
				log.Printf("[ClickHouse] 刷新缓冲区失败: %v", err)
			}
		case <-bw.done:
			return
		}
	}
}

// flush 刷新缓冲区到ClickHouse
func (bw *BatchWriter) flush() error {
	bw.mu.Lock()
	if len(bw.buffer) == 0 {
		bw.mu.Unlock()
		return nil
	}

	// 复制缓冲区数据
	data := make([]GameLogDetail, len(bw.buffer))
	copy(data, bw.buffer)
	bw.buffer = bw.buffer[:0]
	bw.mu.Unlock()

	// 批量写入（带重试）
	var err error
	for i := 0; i <= bw.config.RetryTimes; i++ {
		if i > 0 {
			time.Sleep(bw.config.RetryInterval * time.Duration(i))
		}

		err = bw.batchInsert(data)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("批量写入失败（重试%d次后）: %w", bw.config.RetryTimes, err)
}

// batchInsert 批量插入数据
func (bw *BatchWriter) batchInsert(data []GameLogDetail) error {
	if len(data) == 0 {
		return nil
	}

	log.Printf("[ClickHouse] 开始批量插入 %d 条记录到表 %s", len(data), bw.tableName)

	ctx, cancel := context.WithTimeout(context.Background(), bw.config.Timeout)
	defer cancel()

	// 使用批量插入API
	batch, err := bw.conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", bw.tableName))
	if err != nil {
		return fmt.Errorf("准备批量插入失败: %w", err)
	}

	for _, entry := range data {
		err := batch.Append(
			entry.LogID,
			entry.GameSessionID,
			entry.IntegratorID,
			entry.UserID,
			entry.GameID,
			entry.BetAmount,
			entry.WinAmount,
			entry.NetResult,
			entry.BetLines,
			entry.BetPerLine,
			entry.IsFreeSpin,
			entry.BonusFeature,
			entry.DeviceType,
			entry.DeviceOS,
			entry.BrowserType,
			entry.IPAddress,
			entry.IPRegion,
			entry.IPCountry,
			entry.SessionID,
			entry.ServerID,
			entry.ProcessingTime,
			entry.ErrorCode,
			entry.ErrorMessage,
			entry.GameResultJSON,
			entry.ReelResult,
			entry.WinLines,
			entry.UserAgent,
			entry.LogTime,
		)
		if err != nil {
			return fmt.Errorf("追加数据失败: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("发送批量数据失败: %w", err)
	}

	log.Printf("[ClickHouse] 批量插入成功 %d 条记录", len(data))
	return nil
}

// Flush 手动刷新缓冲区
func (bw *BatchWriter) Flush() error {
	return bw.flush()
}

// Close 关闭写入器
func (bw *BatchWriter) Close() error {
	bw.flushTicker.Stop()
	close(bw.done)

	// 刷新剩余数据
	if err := bw.Flush(); err != nil {
		return err
	}

	return bw.conn.Close()
}

// GetBufferSize 获取当前缓冲区大小
func (bw *BatchWriter) GetBufferSize() int {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	return len(bw.buffer)
}

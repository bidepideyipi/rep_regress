package db

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/rtp-processor/config"
	"github.com/rtp-processor/models"
)

// ClickHouseWriter ClickHouse写入器
type ClickHouseWriter struct {
	conn        clickhouse.Conn
	mutex       sync.Mutex
	buffer      []models.GameLogDetail
	batchSize   int
	tableName   string
	flushCount  int64
	errorCount  int64
}

// NewClickHouseWriter 创建ClickHouse写入器
func NewClickHouseWriter(cfg *config.Config) (*ClickHouseWriter, error) {
	chConfig := cfg.ClickHouse
	if chConfig.Host == "" || chConfig.Port == 0 || chConfig.Username == "" || chConfig.Database == "" {
		return nil, fmt.Errorf("ClickHouse配置不完整")
	}

	batchSize := 100
	if cfg.RocketMQ.Consumer.BatchSize > 0 {
		batchSize = cfg.RocketMQ.Consumer.BatchSize
	}

	log.Printf("[ClickHouse] 初始化写入器: %s:%d/%s, batchSize=%d",
		chConfig.Host, chConfig.Port, chConfig.Database, batchSize)

	dsn := fmt.Sprintf("%s:%d", chConfig.Host, chConfig.Port)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dsn},
		Auth: clickhouse.Auth{
			Database: chConfig.Database,
			Username: chConfig.Username,
			Password: chConfig.Password,
		},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, err
	}

	log.Printf("[ClickHouse] 连接成功")

	return &ClickHouseWriter{
		conn:      conn,
		buffer:    make([]models.GameLogDetail, 0, batchSize),
		batchSize: batchSize,
		tableName: "game_log_detail_local",
	}, nil
}

// AddToBuffer 添加日志到缓冲区
func (w *ClickHouseWriter) AddToBuffer(entry models.GameLogDetail) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.buffer = append(w.buffer, entry)

	if len(w.buffer) >= w.batchSize {
		return w.flushLocked()
	}

	return nil
}

// flushLocked 刷新缓冲区（内部方法，调用前必须持有锁）
func (w *ClickHouseWriter) flushLocked() error {
	if len(w.buffer) == 0 {
		return nil
	}

	if w.conn == nil {
		log.Printf("[ClickHouse] 未连接，清空缓冲区")
		w.buffer = w.buffer[:0]
		return nil
	}

	log.Printf("[ClickHouse] 批量写入 %d 条记录", len(w.buffer))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	batch, err := w.conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", w.tableName))
	if err != nil {
		atomic.AddInt64(&w.errorCount, 1)
		log.Printf("[ClickHouse] 准备批量插入失败: %v", err)
		return err
	}

	for _, entry := range w.buffer {
		err := batch.Append(
			entry.LogID, entry.GameSessionID, entry.IntegratorID, entry.UserID,
			entry.GameID, entry.BetAmount, entry.WinAmount, entry.NetResult,
			entry.BetLines, entry.BetPerLine, entry.IsFreeSpin, entry.BonusFeature,
			entry.DeviceType, entry.DeviceOS, entry.BrowserType, entry.IPAddress,
			entry.IPRegion, entry.IPCountry, entry.SessionID, entry.ServerID,
			entry.ProcessingTime, entry.ErrorCode, entry.ErrorMessage,
			entry.GameResultJSON, entry.ReelResult, entry.WinLines,
			entry.UserAgent, entry.LogTime,
		)
		if err != nil {
			atomic.AddInt64(&w.errorCount, 1)
			log.Printf("[ClickHouse] 追加数据失败: %v", err)
			return err
		}
	}

	if err := batch.Send(); err != nil {
		atomic.AddInt64(&w.errorCount, 1)
		log.Printf("[ClickHouse] 发送批量数据失败: %v", err)
		return err
	}

	w.buffer = w.buffer[:0]
	atomic.AddInt64(&w.flushCount, 1)
	log.Printf("[ClickHouse] 批量写入成功，总成功次数: %d", atomic.LoadInt64(&w.flushCount))

	return nil
}

// Flush 刷新缓冲区
func (w *ClickHouseWriter) Flush() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.flushLocked()
}

// Close 关闭写入器
func (w *ClickHouseWriter) Close() error {
	log.Printf("[ClickHouse] 关闭写入器")

	if w.conn != nil {
		if len(w.buffer) > 0 {
			log.Printf("[ClickHouse] 关闭前刷新剩余 %d 条数据", len(w.buffer))
			w.Flush()
		}
		return w.conn.Close()
	}
	return nil
}

// GetStats 获取统计信息
func (w *ClickHouseWriter) GetStats() (flushCount, errorCount int64, bufferSize int) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return atomic.LoadInt64(&w.flushCount),
		atomic.LoadInt64(&w.errorCount),
		len(w.buffer)
}

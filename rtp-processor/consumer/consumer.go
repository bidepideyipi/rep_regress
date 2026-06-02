package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/rtp-processor/config"
	"github.com/rtp-processor/models"
)

// NewGameConsumer 创建游戏消费者
func NewGameConsumer(appConfig *config.Config) (*GameConsumer, error) {
	gc := &GameConsumer{
		appConfig: appConfig,
	}

	// 初始化 ClickHouse Writer
	chWriter, err := NewClickHouseWriter(appConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化ClickHouse失败: %v", err)
	}
	gc.chWriter = chWriter

	// 初始化 RocketMQ Consumer
	if err := gc.initRocketMQConsumer(); err != nil {
		chWriter.Close()
		return nil, err
	}

	return gc, nil
}

// NewClickHouseWriter 创建ClickHouseWriter
func NewClickHouseWriter(appConfig *config.Config) (*ClickHouseWriter, error) {
	chConfig := appConfig.ClickHouse
	if chConfig.Host == "" || chConfig.Port == 0 || chConfig.Username == "" || chConfig.Database == "" {
		log.Printf("[Consumer] ClickHouseHouse配置不完整，无法初始化ClickHouseWriter")
		return nil, fmt.Errorf("ClickHouseHouse配置不完整，无法初始化ClickHouseWriter")
	}

	batchSize := 100
	if appConfig.RocketMQ.Consumer.BatchSize > 0 {
		batchSize = appConfig.RocketMQ.Consumer.BatchSize
	}

	log.Printf("[Consumer] 初始化 ClickHouseWriter, host=%s:%d, batchSize=%d",
		chConfig.Host, chConfig.Port, batchSize)

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

	log.Printf("[Consumer] ClickHouse连接成功")

	writer := &ClickHouseWriter{
		conn:      conn,
		buffer:    make([]models.GameLogDetail, 0, batchSize),
		batchSize: batchSize,
		tableName: "game_log_detail_local",
	}

	return writer, nil
}

// AddToBuffer 添加日志到缓冲区（同步方法）
func (w *ClickHouseWriter) AddToBuffer(entry models.GameLogDetail) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.buffer = append(w.buffer, entry)

	// 达到批量大小时立即同步刷新
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
		log.Printf("[Consumer] ClickHouse未连接，清空缓冲区")
		w.buffer = w.buffer[:0]
		return nil
	}

	log.Printf("[Consumer] 开始批量写入 %d 条记录", len(w.buffer))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	batch, err := w.conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", w.tableName))
	if err != nil {
		atomic.AddInt64(&w.errorCount, 1)
		log.Printf("[Consumer] 准备批量插入失败: %v", err)
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
			log.Printf("[Consumer] 追加数据失败: %v", err)
			return err
		}
	}

	if err := batch.Send(); err != nil {
		atomic.AddInt64(&w.errorCount, 1)
		log.Printf("[Consumer] 发送批量数据失败: %v", err)
		return err
	}

	w.buffer = w.buffer[:0]
	atomic.AddInt64(&w.flushCount, 1)
	log.Printf("[Consumer] 批量写入成功，总成功次数: %d", atomic.LoadInt64(&w.flushCount))

	return nil
}

// Flush 刷新缓冲区（外部调用，同步方法）
func (w *ClickHouseWriter) Flush() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.flushLocked()
}

// Close 关闭写入器
func (w *ClickHouseWriter) Close() error {
	log.Printf("[Consumer] 关闭写入器")

	if w.conn != nil {
		if len(w.buffer) > 0 {
			log.Printf("[Consumer] 关闭前刷新剩余 %d 条数据", len(w.buffer))
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

func (gc *GameConsumer) parseRocketMQNameServers() string {
	if len(gc.appConfig.RocketMQ.NameServers) == 0 {
		return "127.0.0.1:9876"
	}
	nsList := make([]string, len(gc.appConfig.RocketMQ.NameServers))
	copy(nsList, gc.appConfig.RocketMQ.NameServers)
	return strings.Join(nsList, ";")
}

func (gc *GameConsumer) getConsumerGroupAndTopic() (string, string) {
	groupName := gc.appConfig.RocketMQ.Consumer.GroupName
	if groupName == "" {
		groupName = "slot_game_consumer_group"
	}

	topic := gc.appConfig.RocketMQ.Consumer.Topic
	if topic == "" {
		topic = gc.appConfig.RocketMQ.Producer.Topic
		if topic == "" {
			topic = "game_log_topic"
		}
	}

	return groupName, topic
}

/**
 * @brief 初始化RocketMQ消费者
 * @return error 初始化失败时返回错误
 */
func (gc *GameConsumer) initRocketMQConsumer() error {
	nameServers := gc.parseRocketMQNameServers()
	groupName, topic := gc.getConsumerGroupAndTopic()

	log.Printf("[Consumer] RocketMQ Pull模式配置: nameServers=%s, groupName=%s, topic=%s",
		nameServers, groupName, topic)

	namesrv, err := primitive.NewNamesrvAddr(nameServers)
	if err != nil {
		return fmt.Errorf("创建Namesrv失败: %v", err)
	}

	rocketConsumer, err := rocketmq.NewPullConsumer(
		consumer.WithGroupName(groupName),
		consumer.WithNameServer(namesrv),
	)
	if err != nil {
		return fmt.Errorf("创建PullConsumer失败: %v", err)
	}

	// 订阅主题
	if err := rocketConsumer.Subscribe(topic, consumer.MessageSelector{}); err != nil {
		rocketConsumer.Shutdown()
		return fmt.Errorf("订阅主题失败: %v", err)
	}

	gc.rocketConsumer = rocketConsumer
	gc.stopChan = make(chan struct{})
	gc.offsets = make(map[int64]int64)
	return nil
}

/**
 * @brief 拉取并处理消息
 * @param ctx 上下文
 */
func (gc *GameConsumer) pullAndProcess(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-gc.stopChan:
			return
		default:
		}

		// Pull 方法批量拉取消息，每次最多 100 条
		pullResult, err := gc.rocketConsumer.Pull(ctx, 100)
		if err != nil {
			log.Printf("[Consumer] Pull 失败: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		msgs := pullResult.GetMessageExts()
		if len(msgs) > 0 {
			log.Printf("[Consumer] 拉取到 %d 条消息 (MinOffset: %d, MaxOffset: %d, NextBegin: %d)",
				len(msgs), pullResult.MinOffset, pullResult.MaxOffset, pullResult.NextBeginOffset)

			// 先添加到缓冲区
			for _, msg := range msgs {
				var logEntry models.GameLogDetail
				if err := json.Unmarshal(msg.Body, &logEntry); err != nil {
					log.Printf("[Consumer] JSON解析失败: %v", err)
					continue
				}

				if err := gc.chWriter.AddToBuffer(logEntry); err != nil {
					log.Printf("[Consumer] 写入失败: %v", err)
				}
			}

			// 刷新到 ClickHouse，只有写入成功才算消费成功
			if err := gc.chWriter.Flush(); err != nil {
				log.Printf("[Consumer] 刷新 ClickHouse 失败，停止消费: %v", err)
				// 写入失败，停止消费，避免数据丢失
				// 注意：offset 未持久化，重启后会重新消费这批消息
				time.Sleep(time.Second)
				continue
			}

			// 写入成功后立即持久化 offset，确保数据不丢失
			_, topic := gc.getConsumerGroupAndTopic()
			if err := gc.rocketConsumer.PersistOffset(ctx, topic); err != nil {
				log.Printf("[Consumer] 持久化 offset 失败: %v", err)
			}

			// 处理限流延迟
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Start 启动消费者
func (gc *GameConsumer) Start() error {
	if err := gc.rocketConsumer.Start(); err != nil {
		return fmt.Errorf("启动消费者失败: %v", err)
	}

	log.Printf("[Consumer] PullConsumer 启动成功，开始拉取消息")

	// 启动拉取循环
	ctx := context.Background()
	go gc.pullAndProcess(ctx)

	return nil
}

// Stop 停止消费者
func (gc *GameConsumer) Stop() error {
	log.Printf("[Consumer] 正在关闭")

	// 停止拉取循环
	if gc.stopChan != nil {
		close(gc.stopChan)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		if gc.rocketConsumer != nil {
			if err := gc.rocketConsumer.Shutdown(); err != nil {
				log.Printf("[Consumer] 关闭消费者失败: %v", err)
			}
		}

		if gc.chWriter != nil {
			if err := gc.chWriter.Flush(); err != nil {
				log.Printf("[Consumer] 刷新剩余数据失败: %v", err)
			}

			flushCount, errorCount, bufferSize := gc.chWriter.GetStats()
			log.Printf("[Consumer] ======== 关闭统计 ========")
			log.Printf("[Consumer] 成功刷新次数: %d", flushCount)
			log.Printf("[Consumer] 失败次数: %d", errorCount)
			log.Printf("[Consumer] 关闭时缓冲区大小: %d", bufferSize)
			log.Printf("[Consumer] =========================")

			if err := gc.chWriter.Close(); err != nil {
				log.Printf("[Consumer] 关闭写入器失败: %v", err)
			}
		}

		close(done)
	}()

	select {
	case <-ctx.Done():
		log.Printf("[Consumer] 关闭超时，强制退出")
		return ctx.Err()
	case <-done:
		log.Printf("[Consumer] 优雅关闭完成")
		return nil
	}
}

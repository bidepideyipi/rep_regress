package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"platform-games/slot-game/models"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// NewGameConsumer 创建游戏消费者
func NewGameConsumer(appConfig *AppConfig) (*GameConsumer, error) {
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
func NewClickHouseWriter(appConfig *AppConfig) (*ClickHouseWriter, error) {
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
	//log.Printf("[Consumer] 缓冲区大小: %d/%d", len(w.buffer), w.batchSize)

	// 达到批量大小时立即同步刷新
	if len(w.buffer) >= w.batchSize {
		log.Printf("[Consumer] 达到批量大小，开始同步刷新")
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

	log.Printf("[Consumer] RocketMQ配置: nameServers=%s, groupName=%s, topic=%s",
		nameServers, groupName, topic)

	namesrv, err := primitive.NewNamesrvAddr(nameServers)
	if err != nil {
		return fmt.Errorf("创建Namesrv失败: %v", err)
	}

	rocketConsumer, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(groupName),
		consumer.WithNameServer(namesrv),
		consumer.WithPullBatchSize(64),
		consumer.WithPullInterval(5*time.Second),
	)
	if err != nil {
		return fmt.Errorf("创建消费者失败: %v", err)
	}

	// 订阅主题
	err = rocketConsumer.Subscribe(topic, consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		log.Printf("[Consumer] 收到 %d 条消息", len(msgs))

		for _, msg := range msgs {
			var logEntry models.GameLogDetail
			if err := json.Unmarshal(msg.Body, &logEntry); err != nil {
				log.Printf("[Consumer] JSON解析失败: %v, body=%s", err, string(msg.Body))
				continue
			}

			log.Printf("[Consumer] 解析成功, log_id=%s, game_id=%s", logEntry.LogID, logEntry.GameID)

			if err := gc.chWriter.AddToBuffer(logEntry); err != nil {
				log.Printf("[Consumer] 写入失败: %v, msg_id=%s, log_id=%s", err, msg.MsgId, logEntry.LogID)
				return consumer.ConsumeRetryLater, err
			}
		}

		// 批量处理完成后刷新到 ClickHouse，成功后再 ack
		if err := gc.chWriter.Flush(); err != nil {
			log.Printf("[Consumer] 刷新 ClickHouse 失败: %v", err)
			return consumer.ConsumeRetryLater, err
		}

		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		return fmt.Errorf("订阅主题失败: %v", err)
	}

	gc.rocketConsumer = rocketConsumer
	return nil
}

// Start 启动消费者
func (gc *GameConsumer) Start() error {
	if err := gc.rocketConsumer.Start(); err != nil {
		return fmt.Errorf("启动消费者失败: %v", err)
	}

	log.Printf("[Consumer] 消费者启动成功")
	return nil
}

// Stop 停止消费者
func (gc *GameConsumer) Stop() error {
	log.Printf("[Consumer] 正在关闭")

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

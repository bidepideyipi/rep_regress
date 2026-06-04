package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/rtp-processor/config"
	"github.com/rtp-processor/db"
	"github.com/rtp-processor/jackpot"
	"github.com/rtp-processor/models"
)

// NewGameConsumer 创建游戏消费者
//
// 该函数初始化一个完整的游戏日志消费者，包括：
// 1. ClickHouse Writer - 用于批量写入游戏日志到 ClickHouse
// 2. MySQL Manager - 用于数据库操作（如 Jackpot 计算）
// 3. Jackpot Manager - 用于累计奖池计算
// 4. RocketMQ PullConsumer - 用于从消息队列拉取消息
//
// 参数:
//   - appConfig: 应用配置，包含 ClickHouse、MySQL、RocketMQ 等连接信息
//
// 返回:
//   - *GameConsumer: 初始化成功的消费者实例
//   - error: 初始化过程中的错误
//
// 依赖顺序（启动顺序）:
// 1. ClickHouse（数据存储）
// 2. MySQL（业务数据）
// 3. RocketMQ NameServer
// 4. RocketMQ Broker
func NewGameConsumer(appConfig *config.Config) (*GameConsumer, error) {
	gc := &GameConsumer{
		appConfig: appConfig,
	}

	// ============ 第一步：初始化 ClickHouse Writer ============
	// ClickHouse 用于存储游戏日志，支持大规模时序数据写入
	chWriter, err := db.NewClickHouseWriter(appConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化ClickHouse失败: %v", err)
	}
	gc.chWriter = chWriter

	// ============ 第二步：初始化 MySQL Manager ============
	// MySQL 用于存储业务数据，如 Jackpot 累计奖池状态
	// 此步骤失败不中断流程，Jackpot 功能会优雅降级
	mysqlMgr, err := db.NewMySQLManager(appConfig)
	if err != nil {
		log.Printf("[Consumer] 初始化MySQLManager失败: %v", err)
	} else {
		gc.mysqlManager = mysqlMgr

		// ============ 第三步：初始化 Jackpot Manager ============
		// Jackpot Manager 负责累计奖池的计算和更新
		// 依赖 MySQL Manager
		jackpotMgr, err := jackpot.NewManager(mysqlMgr)
		if err != nil {
			log.Printf("[Consumer] 初始化JackpotManager失败: %v", err)
		} else {
			gc.jackpotManager = jackpotMgr
		}
	}

	// ============ 第四步：初始化 RocketMQ Consumer ============
	// 此步骤失败需要清理之前创建的资源
	if err := gc.initRocketMQConsumer(); err != nil {
		// 资源清理：按相反顺序关闭已初始化的组件
		chWriter.Close()
		if gc.jackpotManager != nil {
			gc.jackpotManager.Close()
		}
		if gc.mysqlManager != nil {
			gc.mysqlManager.Close()
		}
		return nil, err
	}

	return gc, nil
}

// parseRocketMQNameServers 解析 RocketMQ NameServer 地址列表
//
// 将配置中的 NameServer 地址数组转换为 RocketMQ 客户端需要的字符串格式
// 多个地址使用分号（;）分隔，例如："127.0.0.1:9876;127.0.0.1:9877"
//
// 返回:
//   - string: 格式化后的 NameServer 地址字符串，如果配置为空则返回默认地址 "127.0.0.1:9876"
func (gc *GameConsumer) parseRocketMQNameServers() string {
	if len(gc.appConfig.RocketMQ.NameServers) == 0 {
		return "127.0.0.1:9876"
	}
	nsList := make([]string, len(gc.appConfig.RocketMQ.NameServers))
	copy(nsList, gc.appConfig.RocketMQ.NameServers)
	return strings.Join(nsList, ";")
}

// getConsumerGroupAndTopic 获取消费者组名和主题名称
//
// 从配置中读取消费者组和主题信息，支持降级到默认值：
// - 消费者组名: 优先使用配置，否则使用默认值 "slot_game_consumer_group"
// - 主题名: 优先使用 consumer.topic，否则使用 producer.topic，最后使用默认值 "game_log_topic"
//
// 返回:
//   - string: 消费者组名
//   - string: 主题名称
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

// initRocketMQConsumer 初始化 RocketMQ Pull 模式消费者
//
// Pull 模式允许消费者主动控制消息拉取的时机和批量大小，
// 相比 Push 模式更适合需要精细控制消费速率的场景。
//
// 初始化步骤:
// 1. 解析 NameServer 地址
// 2. 创建 PullConsumer 实例
// 3. 订阅指定主题
// 4. 初始化偏移量记录和停止信号通道
//
// 返回:
//   - error: 初始化过程中的错误
func (gc *GameConsumer) initRocketMQConsumer() error {
	// 设置 RocketMQ 客户端日志级别为 Error，减少输出
	rlog.SetLogLevel("error")

	nameServers := gc.parseRocketMQNameServers()
	groupName, topic := gc.getConsumerGroupAndTopic()

	log.Printf("[Consumer] RocketMQ Pull模式: nameServers=%s, groupName=%s, topic=%s",
		nameServers, groupName, topic)

	// 创建 NameServer 地址解析器
	namesrv, err := primitive.NewNamesrvAddr(nameServers)
	if err != nil {
		return fmt.Errorf("创建Namesrv失败: %v", err)
	}

	// 创建 Pull 模式消费者
	rocketConsumer, err := rocketmq.NewPullConsumer(
		consumer.WithGroupName(groupName),
		consumer.WithNameServer(namesrv),
		consumer.WithPullBatchSize(100),
		consumer.WithConsumerPullTimeout(3*time.Second), // Pull 超时 3 秒
	)
	if err != nil {
		return fmt.Errorf("创建PullConsumer失败: %v", err)
	}

	// 订阅主题，MessageSelector{} 表示订阅所有消息
	if err := rocketConsumer.Subscribe(topic, consumer.MessageSelector{}); err != nil {
		rocketConsumer.Shutdown()
		return fmt.Errorf("订阅主题失败: %v", err)
	}

	gc.rocketConsumer = rocketConsumer
	gc.stopChan = make(chan struct{})
	gc.offsets = make(map[int64]int64)
	return nil
}

// pullAndProcess 拉取并处理消息的主循环
//
// 该方法在一个独立的 goroutine 中运行，持续执行以下流程：
// 1. 从 RocketMQ 拉取消息（批量）
// 2. 解析消息体为 GameLogDetail 结构
// 3. 将数据写入 ClickHouse 缓冲区
// 4. 计算 Jackpot 累计奖池
// 5. 刷新 ClickHouse 缓冲区（确保数据写入）
// 6. 持久化消费偏移量
//
// 错误处理:
// - 拉取失败：等待 500ms 后重试
// - 刷新失败：等待 1s 后重试（不处理该批次消息）
// - JSON 解析失败：跳过该消息，继续处理下一条
//
// 参数:
//   - ctx: 上下文，用于支持取消操作
func (gc *GameConsumer) pullAndProcess(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// 上下文取消，退出循环
			return
		case <-gc.stopChan:
			// 接收到停止信号，退出循环
			return
		default:
			// 继续执行
		}

		// 从 RocketMQ 拉取消息，最多 100 条
		// 100 只是"期望上限"，实际返回由 Broker 的 pullBatchSize 决定。
		// 实际上是 看broker.conf 中的 pullBatchSize 条消息（默认 32）。
		beforePull := time.Now()
		pullCtx, pullCancel := context.WithTimeout(ctx, 3*time.Second)
		pullResult, err := gc.rocketConsumer.Pull(pullCtx, 100)
		pullCancel()
		pullDuration := time.Since(beforePull)
		if pullDuration > time.Second {
			log.Printf("[Consumer] Pull 耗时: %v", pullDuration)
		}
		if err != nil {
			if err != context.DeadlineExceeded {
				log.Printf("[Consumer] Pull 失败: %v", err)
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		msgs := pullResult.GetMessageExts()
		if len(msgs) > 0 {
			log.Printf("[Consumer] 拉取到 %d 条消息", len(msgs))

			// 处理每条消息
			for _, msg := range msgs {
				var logEntry models.GameLogDetail
				if err := json.Unmarshal(msg.Body, &logEntry); err != nil {
					log.Printf("[Consumer] JSON解析失败: %v", err)
					continue
				}

				// 写入 ClickHouse 缓冲区
				if err := gc.chWriter.AddToBuffer(logEntry); err != nil {
					log.Printf("[Consumer] 写入失败: %v", err)
				}

				// 计算 Jackpot 累计奖池
				// if gc.jackpotManager != nil {
				// 	gc.jackpotManager.CalculateJackpot(&logEntry)
				// }
			}

			// 刷新 ClickHouse 缓冲区，确保数据写入
			if err := gc.chWriter.Flush(); err != nil {
				log.Printf("[Consumer] 刷新失败: %v", err)
				time.Sleep(time.Second)
				continue
			}

			// 持久化消费偏移量，确保消息不丢失
			_, topic := gc.getConsumerGroupAndTopic()
			if err := gc.rocketConsumer.PersistOffset(ctx, topic); err != nil {
				log.Printf("[Consumer] 持久化offset失败: %v", err)
			}

			// 可选：控制拉取频率
			//time.Sleep(100 * time.Millisecond)
		}
	}
}

// Start 启动消费者
//
// 启动 RocketMQ 消费者并开始拉取消息处理。
// 该方法会在新的 goroutine 中运行 pullAndProcess 主循环，
// 立即返回，不阻塞调用线程。
//
// 返回:
//   - error: 启动过程中的错误
func (gc *GameConsumer) Start() error {
	// 启动 RocketMQ 消费者
	if err := gc.rocketConsumer.Start(); err != nil {
		return fmt.Errorf("启动消费者失败: %v", err)
	}

	log.Printf("[Consumer] PullConsumer 启动成功")

	// 在新的 goroutine 中启动消息处理主循环
	ctx := context.Background()
	go gc.pullAndProcess(ctx)

	return nil
}

// Stop 停止消费者
//
// 优雅关闭消费者，执行以下步骤：
// 1. 发送停止信号，中断消息拉取循环
// 2. 关闭 RocketMQ 消费者
// 3. 关闭 Jackpot Manager
// 4. 关闭 MySQL Manager
// 5. 刷新并关闭 ClickHouse Writer
// 6. 输出关闭统计信息
//
// 该方法会阻塞最多 30 秒，等待所有资源关闭完成。
// 超时后会返回 context.DeadlineExceeded 错误。
//
// 返回:
//   - error: 关闭过程中的错误
func (gc *GameConsumer) Stop() error {
	log.Printf("[Consumer] 正在关闭")

	// 发送停止信号
	if gc.stopChan != nil {
		close(gc.stopChan)
	}

	// 设置 30 秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		// 关闭 RocketMQ 消费者
		if gc.rocketConsumer != nil {
			if err := gc.rocketConsumer.Shutdown(); err != nil {
				log.Printf("[Consumer] 关闭消费者失败: %v", err)
			}
		}

		// 关闭 Jackpot Manager
		if gc.jackpotManager != nil {
			gc.jackpotManager.Close()
		}

		// 关闭 MySQL Manager
		if gc.mysqlManager != nil {
			gc.mysqlManager.Close()
		}

		// 关闭 ClickHouse Writer
		if gc.chWriter != nil {
			// 获取并输出统计信息
			flushCount, errorCount, bufferSize := gc.chWriter.GetStats()
			log.Printf("[Consumer] ======== 关闭统计 ========")
			log.Printf("[Consumer] 成功刷新: %d, 失败: %d, 缓冲: %d",
				flushCount, errorCount, bufferSize)
			log.Printf("[Consumer] =========================")

			gc.chWriter.Close()
		}

		close(done)
	}()

	select {
	case <-ctx.Done():
		// 超时
		return ctx.Err()
	case <-done:
		// 完成
		log.Printf("[Consumer] 优雅关闭完成")
		return nil
	}
}

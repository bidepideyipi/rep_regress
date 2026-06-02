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
	"github.com/rtp-processor/config"
	"github.com/rtp-processor/db"
	"github.com/rtp-processor/jackpot"
	"github.com/rtp-processor/models"
)

// NewGameConsumer 创建游戏消费者
func NewGameConsumer(appConfig *config.Config) (*GameConsumer, error) {
	gc := &GameConsumer{
		appConfig: appConfig,
	}

	// 初始化 ClickHouse Writer
	chWriter, err := db.NewClickHouseWriter(appConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化ClickHouse失败: %v", err)
	}
	gc.chWriter = chWriter

	// 初始化 MySQL Manager
	mysqlMgr, err := db.NewMySQLManager(appConfig)
	if err != nil {
		log.Printf("[Consumer] 初始化MySQLManager失败: %v", err)
	} else {
		gc.mysqlManager = mysqlMgr

		// 初始化 Jackpot Manager
		jackpotMgr, err := jackpot.NewManager(mysqlMgr)
		if err != nil {
			log.Printf("[Consumer] 初始化JackpotManager失败: %v", err)
		} else {
			gc.jackpotManager = jackpotMgr
		}
	}

	// 初始化 RocketMQ Consumer
	if err := gc.initRocketMQConsumer(); err != nil {
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

func (gc *GameConsumer) initRocketMQConsumer() error {
	nameServers := gc.parseRocketMQNameServers()
	groupName, topic := gc.getConsumerGroupAndTopic()

	log.Printf("[Consumer] RocketMQ Pull模式: nameServers=%s, groupName=%s, topic=%s",
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

	if err := rocketConsumer.Subscribe(topic, consumer.MessageSelector{}); err != nil {
		rocketConsumer.Shutdown()
		return fmt.Errorf("订阅主题失败: %v", err)
	}

	gc.rocketConsumer = rocketConsumer
	gc.stopChan = make(chan struct{})
	gc.offsets = make(map[int64]int64)
	return nil
}

func (gc *GameConsumer) pullAndProcess(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-gc.stopChan:
			return
		default:
		}

		pullResult, err := gc.rocketConsumer.Pull(ctx, 100)
		if err != nil {
			log.Printf("[Consumer] Pull 失败: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		msgs := pullResult.GetMessageExts()
		if len(msgs) > 0 {
			log.Printf("[Consumer] 拉取到 %d 条消息", len(msgs))

			for _, msg := range msgs {
				var logEntry models.GameLogDetail
				if err := json.Unmarshal(msg.Body, &logEntry); err != nil {
					log.Printf("[Consumer] JSON解析失败: %v", err)
					continue
				}

				if err := gc.chWriter.AddToBuffer(logEntry); err != nil {
					log.Printf("[Consumer] 写入失败: %v", err)
				}

				if gc.jackpotManager != nil {
					gc.jackpotManager.CalculateJackpot(&logEntry)
				}
			}

			if err := gc.chWriter.Flush(); err != nil {
				log.Printf("[Consumer] 刷新失败: %v", err)
				time.Sleep(time.Second)
				continue
			}

			_, topic := gc.getConsumerGroupAndTopic()
			if err := gc.rocketConsumer.PersistOffset(ctx, topic); err != nil {
				log.Printf("[Consumer] 持久化offset失败: %v", err)
			}

			//time.Sleep(100 * time.Millisecond)
		}
	}
}

func (gc *GameConsumer) Start() error {
	if err := gc.rocketConsumer.Start(); err != nil {
		return fmt.Errorf("启动消费者失败: %v", err)
	}

	log.Printf("[Consumer] PullConsumer 启动成功")

	ctx := context.Background()
	go gc.pullAndProcess(ctx)

	return nil
}

func (gc *GameConsumer) Stop() error {
	log.Printf("[Consumer] 正在关闭")

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

		if gc.jackpotManager != nil {
			gc.jackpotManager.Close()
		}

		if gc.mysqlManager != nil {
			gc.mysqlManager.Close()
		}

		if gc.chWriter != nil {
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
		return ctx.Err()
	case <-done:
		log.Printf("[Consumer] 优雅关闭完成")
		return nil
	}
}

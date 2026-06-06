package consumer

import (
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
)

// NewPushConsumer 创建Push模式消费者
func NewPushConsumer(appConfig *config.Config) (*PushConsumer, error) {
	pc := &PushConsumer{
		appConfig: appConfig,
		stopChan:  make(chan struct{}),
	}

	// 初始化 ClickHouse Writer
	chWriter, err := db.NewClickHouseWriter(appConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化ClickHouse失败: %v", err)
	}

	// 初始化 MySQL Manager
	var mysqlMgr *db.MySQLManager
	var jackpotMgr *jackpot.Manager
	mysqlMgr, err = db.NewMySQLManager(appConfig)
	if err != nil {
		log.Printf("[PushConsumer] 初始化MySQLManager失败: %v", err)
	} else {
		pc.mysqlManager = mysqlMgr

		// 初始化 Jackpot Manager
		jackpotMgr, err = jackpot.NewManager(mysqlMgr)
		if err != nil {
			log.Printf("[PushConsumer] 初始化JackpotManager失败: %v", err)
		}
	}

	// 创建消息监听器（业务逻辑层）
	pc.listener = NewMessageListener(chWriter, jackpotMgr)
	pc.chWriter = chWriter
	pc.jackpotManager = jackpotMgr

	if err := pc.initRocketMQConsumer(); err != nil {
		// 清理已初始化的资源
		chWriter.Close()
		if jackpotMgr != nil {
			jackpotMgr.Close()
		}
		if mysqlMgr != nil {
			mysqlMgr.Close()
		}
		return nil, err
	}

	return pc, nil
}

func (pc *PushConsumer) parseRocketMQNameServers() string {
	if len(pc.appConfig.RocketMQ.NameServers) == 0 {
		return "127.0.0.1:9876"
	}
	return strings.Join(pc.appConfig.RocketMQ.NameServers, ";")
}

func (pc *PushConsumer) getConsumerGroupAndTopic() (string, string) {
	groupName := pc.appConfig.RocketMQ.Consumer.GroupName
	if groupName == "" {
		groupName = "slot_game_consumer_group"
	}

	topic := pc.appConfig.RocketMQ.Consumer.Topic
	if topic == "" {
		topic = pc.appConfig.RocketMQ.Producer.Topic
		if topic == "" {
			topic = "game_log_topic"
		}
	}

	return groupName, topic
}

func (pc *PushConsumer) initRocketMQConsumer() error {
	nameServers := pc.parseRocketMQNameServers()
	groupName, topic := pc.getConsumerGroupAndTopic()

	log.Printf("[PushConsumer] nameServers=%s, groupName=%s, topic=%s",
		nameServers, groupName, topic)

	// 创建 NameServer 地址解析器
	namesrv, err := primitive.NewNamesrvAddr(nameServers)
	if err != nil {
		return fmt.Errorf("创建Namesrv失败: %v", err)
	}

	// 创建Push模式消费者
	rocketConsumer, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(groupName),
		consumer.WithNameServer(namesrv),
		consumer.WithConsumeMessageBatchMaxSize(100),    // 单次批量消费最大数量
		consumer.WithConsumeTimeout(time.Minute),        // 消费超时
		consumer.WithMaxReconsumeTimes(3),               // 最大重试次数
		consumer.WithConsumerModel(consumer.Clustering), // 集群模式
	)
	if err != nil {
		return fmt.Errorf("创建PushConsumer失败: %v", err)
	}

	// 注册消息监听器（并发消费）
	err = rocketConsumer.Subscribe(topic, consumer.MessageSelector{}, pc.listener.Consume)
	if err != nil {
		rocketConsumer.Shutdown()
		return fmt.Errorf("订阅主题失败: %v", err)
	}

	pc.rocketConsumer = rocketConsumer
	return nil
}

// Start 启动消费者
func (pc *PushConsumer) Start() error {
	if err := pc.rocketConsumer.Start(); err != nil {
		return fmt.Errorf("启动消费者失败: %v", err)
	}

	log.Printf("[PushConsumer] 启动成功")
	return nil
}

// Stop 停止消费者
func (pc *PushConsumer) Stop() error {
	log.Printf("[PushConsumer] 正在关闭")

	close(pc.stopChan)

	if pc.rocketConsumer != nil {
		if err := pc.rocketConsumer.Shutdown(); err != nil {
			log.Printf("[PushConsumer] 关闭失败: %v", err)
			return err
		}
	}

	// 关闭 Jackpot Manager
	if pc.jackpotManager != nil {
		pc.jackpotManager.Close()
	}

	// 关闭 MySQL Manager
	if pc.mysqlManager != nil {
		pc.mysqlManager.Close()
	}

	// 关闭 ClickHouse Writer
	if pc.chWriter != nil {
		flushCount, errorCount := pc.chWriter.GetStats()
		log.Printf("[PushConsumer] ======== 关闭统计 ========")
		log.Printf("[PushConsumer] 成功刷新: %d, 失败: %d", flushCount, errorCount)
		log.Printf("[PushConsumer] =========================")
		pc.chWriter.Close()
	}

	log.Printf("[PushConsumer] 优雅关闭完成")
	return nil
}

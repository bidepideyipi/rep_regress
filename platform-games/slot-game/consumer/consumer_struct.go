package consumer

import (
	"sync"

	"platform-games/slot-game/models"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/apache/rocketmq-client-go/v2"
)

// AppConfig 应用配置结构
type AppConfig struct {
	RocketMQ   RocketMQConfig   `json:"rocket_mq"`
	ClickHouse ClickHouseConfig `json:"clickhouse"`
}

// RocketMQConfig RocketMQ配置
type RocketMQConfig struct {
	NameServers []string       `json:"name_servers"`
	Producer    ProducerConfig `json:"producer"`
	Consumer    ConsumerConfig `json:"consumer"`
	ACL         ACLConfig      `json:"acl,omitempty"`
}

// ProducerConfig 生产者配置
type ProducerConfig struct {
	GroupName  string `json:"group_name"`
	Topic      string `json:"topic"`
	Timeout    int    `json:"timeout"`
	RetryTimes int    `json:"retry_times"`
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	GroupName   string `json:"group_name"`
	Topic       string `json:"topic"`
	BatchSize   int    `json:"batch_size"`
	ThreadCount int    `json:"thread_count"`
}

// ACLConfig ACL配置
type ACLConfig struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

// ClickHouseConfig ClickHouse配置
type ClickHouseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// ClickHouseWriter ClickHouse批量写入器
type ClickHouseWriter struct {
	conn       clickhouse.Conn
	buffer     []models.GameLogDetail
	mutex      sync.Mutex
	batchSize  int
	flushCount int64
	errorCount int64
	tableName  string
}

// GameConsumer 游戏消费者
type GameConsumer struct {
	appConfig      *AppConfig
	chWriter       *ClickHouseWriter
	rocketConsumer rocketmq.PushConsumer
}
